/*

const { MQ } = require('./mq');

async function main() {

    try {
        const mq = await MQ.connect('mq://user:pass@localhost:8080');
        //("test.t555.test"
        mq.sub("test.*.test", (data) => {
            console.log('Received:', data.payload);
        });
        mq.sub('topic', (data) => {
            console.log('Received:', data.payload);
        });
                mq.pub('topic', 'Hello World');
        mq.service('service', (data, reply) => {
            reply("",'Response to: ' + data.payload);
        });
        
        const response = await mq.req('service', 'test');
        console.log('Service response:', response);
        

        
    } catch (err) {
        console.error('Error:', err);
    }
}

main();

*/

const { MQ } = require('./mq-client');

async function main() {
    try {
        // Connect to the server
        const mq = await MQ.connect("mq://root:fffffffffffffffffff@localhost:4051");
   console.log('Conectado com ID:', mq.id);

        // Assinar um tópico
        mq.subscribe('test.topic', (msg) => {
            console.log('Mensagem recebida:', msg.payload);
        });

        // Publicar uma mensagem
        await mq.publish('test.topic', 'Olá Mundo!');

        mq.service("test.teste",(data,reply)=>{
            console.log(data)
            reply("dddd",{"nome": "João"})
        })

        try {
       const res= await mq.request("test.teste",{"payload": "dsdsdsdsdsdsdsdsdsdsdsdsdsdsds"},2000)
       console.log(res)
        } catch (error) {
             console.log(error)
        }



        /*
        // Usar KV store
        const kv = mq.kv('meu-bucket');
        await kv.set('chave1', 'valor1');
        const valor = await kv.get('chave1');
        console.log('Valor obtido:', valor);

        // Usar banco de dados
        const db = mq.dbCollection('minha-colecao');
        const id = await db.insert({ nome: "Exemplo", valor: 123 });
        console.log('Documento inserido com ID:', id);

        // Fechar conexão
        //await mq.stop();
        */
    } catch (err) {
        console.error('Error:', err);
    }
}

main();