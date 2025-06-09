package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrDocumentNotFound   = errors.New("document not found")
	ErrInvalidQuery       = errors.New("invalid query")
)

// Document representa um documento no banco de dados (similar ao MongoDB)
type Document map[string]interface{}

// NoSQL é a estrutura principal do nosso banco de dados
type NoSQL struct {
	db      *bbolt.DB
	indexes map[string]*IndexManager
	mu      sync.RWMutex
}

// IndexManager gerencia índices para consultas rápidas
type IndexManager struct {
	indexes map[string]*Index
	mu      sync.RWMutex
}

// Index representa um índice em um campo específico
type Index struct {
	Field string
	Items map[interface{}][]string // Valor -> IDs de documentos
}

// New cria uma nova instância do NoSQL
func New(path string) (*NoSQL, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("kv_store"))
		if err != nil {
			return err
		}
		return nil

	})
	return &NoSQL{
		db:      db,
		indexes: make(map[string]*IndexManager),
	}, nil
}

// Close fecha a conexão com o banco de dados
func (mc *NoSQL) Close() error {
	return mc.db.Close()
}

// NewIndexManager cria um novo gerenciador de índices
func NewIndexManager() *IndexManager {
	return &IndexManager{
		indexes: make(map[string]*Index),
	}
}

// Insert insere um documento na coleção especificada
func (mc *NoSQL) Insert(collection string, doc Document) (string, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var id string
	err := mc.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return err
		}

		id = generateID()
		doc["_id"] = id
		doc["_collection"] = collection

		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		err = b.Put([]byte(id), data)
		if err != nil {
			return err
		}

		// Atualizar índices
		if im, exists := mc.indexes[collection]; exists {
			im.UpdateIndexes(id, doc)
		}

		return nil
	})

	return id, err
}

// FindOne busca um documento por ID
func (mc *NoSQL) FindOne(collection, id string) (Document, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var doc Document
	err := mc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return ErrCollectionNotFound
		}

		data := b.Get([]byte(id))
		if data == nil {
			return ErrDocumentNotFound
		}

		return json.Unmarshal(data, &doc)
	})

	return doc, err
}

// FindAll retorna todos os documentos de uma coleção
func (mc *NoSQL) FindAll(collection string) ([]Document, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var docs []Document
	err := mc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return ErrCollectionNotFound
		}

		return b.ForEach(func(k, v []byte) error {
			var doc Document
			if err := json.Unmarshal(v, &doc); err != nil {
				return err
			}
			docs = append(docs, doc)
			return nil
		})
	})

	return docs, err
}

// Update atualiza um documento existente
func (mc *NoSQL) Update(collection, id string, updates Document) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	err := mc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return ErrCollectionNotFound
		}

		data := b.Get([]byte(id))
		if data == nil {
			return ErrDocumentNotFound
		}

		var doc Document
		if err := json.Unmarshal(data, &doc); err != nil {
			return err
		}

		// Aplicar atualizações
		for k, v := range updates {
			if k != "_id" && k != "_collection" {
				doc[k] = v
			}
		}

		newData, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		err = b.Put([]byte(id), newData)
		if err != nil {
			return err
		}

		// Atualizar índices
		if im, exists := mc.indexes[collection]; exists {
			im.UpdateIndexes(id, doc)
		}

		return nil
	})

	return err
}

// Delete remove um documento
func (mc *NoSQL) Delete(collection, id string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	err := mc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return ErrCollectionNotFound
		}

		// Antes de deletar, precisamos do documento para atualizar índices
		data := b.Get([]byte(id))
		if data == nil {
			return ErrDocumentNotFound
		}

		var doc Document
		if err := json.Unmarshal(data, &doc); err != nil {
			return err
		}

		// Atualizar índices
		if im, exists := mc.indexes[collection]; exists {
			im.RemoveFromIndexes(id, doc)
		}

		return b.Delete([]byte(id))
	})

	return err
}

// CreateIndex cria um índice para um campo específico
func (mc *NoSQL) CreateIndex(collection, field string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if _, exists := mc.indexes[collection]; !exists {
		mc.indexes[collection] = NewIndexManager()
	}
	mc.indexes[collection].CreateIndex(collection, field)
	go func(name string) {
		// Indexar documentos existentes
		docs, err := mc.FindAll(collection)
		if err != nil {
			return
		}

		for _, doc := range docs {
			if id, ok := doc["_id"].(string); ok {
				mc.indexes[collection].UpdateIndexes(id, doc)
			}
		}

	}(collection)

	return nil
}
func (mc *NoSQL) RemoveIndex(collection string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if _, exists := mc.indexes[collection]; !exists {
		delete(mc.indexes, collection)
		return mc.db.Update(func(tx *bbolt.Tx) error {
			err := tx.DeleteBucket([]byte(collection))
			if err != nil {
				return err
			}
			return nil

		})
	}

	return nil
}

// FindWithQuery busca documentos com base em critérios complexos
func (mc *NoSQL) FindWithQuery(collection string, query map[string]interface{}) ([]Document, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Tentar usar índices primeiro
	if ids := mc.tryUseIndex(collection, query); ids != nil {
		return mc.getDocumentsByIds(collection, ids)
	}

	// Caso contrário, verificação completa
	var results []Document
	err := mc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return ErrCollectionNotFound
		}

		return b.ForEach(func(k, v []byte) error {
			var doc Document
			if err := json.Unmarshal(v, &doc); err != nil {
				return err
			}

			if matchQuery(doc, query) {
				results = append(results, doc)
			}

			return nil
		})
	})

	return results, err
}

// tryUseIndex tenta usar índices para otimizar a consulta
func (mc *NoSQL) tryUseIndex(collection string, query map[string]interface{}) []string {
	if im, exists := mc.indexes[collection]; exists {
		for field, value := range query {
			if ids := im.Query(collection, field, value); ids != nil {
				return ids
			}
		}
	}
	return nil
}

// getDocumentsByIds retorna documentos com base em seus IDs
func (mc *NoSQL) getDocumentsByIds(collection string, ids []string) ([]Document, error) {
	var results []Document
	for _, id := range ids {
		doc, err := mc.FindOne(collection, id)
		if err != nil {
			return nil, err
		}
		results = append(results, doc)
	}
	return results, nil
}

// matchQuery verifica se um documento corresponde à consulta
func matchQuery(doc Document, query map[string]interface{}) bool {
	for field, condition := range query {
		docValue, exists := doc[field]
		if !exists {
			return false
		}

		switch cond := condition.(type) {
		case map[string]interface{}: // Operadores como {$gt, $lt, etc.}
			for op, opValue := range cond {
				if !compareWithOperator(docValue, op, opValue) {
					return false
				}
			}
		default: // Comparação direta
			if !reflect.DeepEqual(docValue, condition) {
				return false
			}
		}
	}
	return true
}

// compareWithOperator compara valores com operadores
func compareWithOperator(value interface{}, op string, opValue interface{}) bool {
	switch op {
	case "$gt":
		return compareNumbers(value, opValue) > 0
	case "$gte":
		return compareNumbers(value, opValue) >= 0
	case "$lt":
		return compareNumbers(value, opValue) < 0
	case "$lte":
		return compareNumbers(value, opValue) <= 0
	case "$ne":
		return !reflect.DeepEqual(value, opValue)
	case "$in":
		return inArray(value, opValue)
	case "$nin":
		return !inArray(value, opValue)
	default:
		return false
	}
}

// compareNumbers compara valores numéricos
func compareNumbers(a, b interface{}) int {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)

	if !aok || !bok {
		return 0
	}

	if af > bf {
		return 1
	} else if af < bf {
		return -1
	}
	return 0
}

// toFloat converte valores para float64 quando possível
func toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// inArray verifica se um valor está em um array
func inArray(value interface{}, array interface{}) bool {
	arr, ok := array.([]interface{})
	if !ok {
		return false
	}

	for _, item := range arr {
		if reflect.DeepEqual(value, item) {
			return true
		}
	}
	return false
}

// generateID gera um ID único para documentos
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Métodos do IndexManager

// CreateIndex cria um novo índice
func (im *IndexManager) CreateIndex(collection, field string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	key := collection + ":" + field
	if _, exists := im.indexes[key]; !exists {
		im.indexes[key] = &Index{
			Field: field,
			Items: make(map[interface{}][]string),
		}
	}
}

// UpdateIndexes atualiza todos os índices com o novo documento
func (im *IndexManager) UpdateIndexes(id string, doc Document) {
	im.mu.Lock()
	defer im.mu.Unlock()

	for key, index := range im.indexes {
		if strings.HasPrefix(key, doc["_collection"].(string)+":") {
			field := index.Field
			if value, exists := doc[field]; exists {
				// Remove o ID dos valores antigos (se houver)
				for val, ids := range index.Items {
					newIDs := make([]string, 0, len(ids))
					for _, existingID := range ids {
						if existingID != id {
							newIDs = append(newIDs, existingID)
						}
					}
					index.Items[val] = newIDs
				}

				// Adiciona ao novo valor
				index.Items[value] = append(index.Items[value], id)
			}
		}
	}
}

// RemoveFromIndexes remove um documento dos índices
func (im *IndexManager) RemoveFromIndexes(id string, doc Document) {
	im.mu.Lock()
	defer im.mu.Unlock()

	for key, index := range im.indexes {
		if strings.HasPrefix(key, doc["_collection"].(string)+":") {
			for val, ids := range index.Items {
				newIDs := make([]string, 0, len(ids))
				for _, existingID := range ids {
					if existingID != id {
						newIDs = append(newIDs, existingID)
					}
				}
				index.Items[val] = newIDs
			}
		}
	}
}

// Query busca documentos usando um índice
func (im *IndexManager) Query(collection, field string, value interface{}) []string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	key := collection + ":" + field
	if index, exists := im.indexes[key]; exists {
		return index.Items[value]
	}
	return nil
}

// Exemplo de uso
func main() {
	// Criar uma instância do banco de dados
	db, err := New("mydb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Criar um índice no campo "name"
	if err := db.CreateIndex("users", "name"); err != nil {
		log.Fatal(err)
	}

	// Inserir documentos
	doc1 := Document{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   28,
	}

	id1, err := db.Insert("users", doc1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted document with ID:", id1)

	// Consultar documentos
	user, err := db.FindOne("users", id1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found user: %+v\n", user)

	// Atualizar documento
	update := Document{"age": 29}
	if err := db.Update("users", id1, update); err != nil {
		log.Fatal(err)
	}

	// Consulta complexa
	query := map[string]interface{}{
		"age": map[string]interface{}{
			"$gt": 25,
		},
	}

	results, err := db.FindWithQuery("users", query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nUsers over 25:")
	for _, res := range results {
		fmt.Printf("%+v\n", res)
	}

	// Deletar documento
	if err := db.Delete("users", id1); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nDeleted user with ID:", id1)
}
