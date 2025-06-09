
const net = require('net');
const crypto = require('crypto');
const { EventEmitter } = require('events');

class MQ extends EventEmitter {
    constructor() {
        super();
        this.subs = {};
        this.chs = {};
        this.chrequest = {};
        this.services = {};
        this.id = '';
        this.buffer = '';
    }

    static async connect(url) {
        const info = MQ.parseMQURL(url);
        if (!info) {
            throw new Error('Invalid URL');
        }

        return new Promise((resolve, reject) => {
            const conn = net.createConnection({
                host: info.host,
                port: info.port
            }, () => {
                const mq = new MQ();
                mq.conn = conn;
                
                conn.on('data', (data) => mq.onData(data));
                conn.on('error', (err) => mq.emit('error', err));
                conn.on('close', () => mq.emit('close'));
                
                mq.connect(info.user, info.pass, 1000)
                    .then(() => resolve(mq))
                    .catch(reject);
            });
        });
    }

    static parseMQURL(url) {
        const match = url.match(/^mq:\/\/(?:([^:]+):([^@]+)@)?([^:]+)(?::(\d+))?$/);
        if (!match) return null;
        
        return {
            user: match[1] || '',
            pass: match[2] || '',
            host: match[3] || 'localhost',
            port: match[4] || '1883'
        };
    }

    generateUUID() {
        return crypto.randomBytes(16).toString('hex');
    }

    async connect(username, password, timeout) {
        const reqId = this.generateUUID();
        this.chrequest[reqId] = { resolve: null, reject: null };
        
        const promise = new Promise((resolve, reject) => {
            this.chrequest[reqId].resolve = resolve;
            this.chrequest[reqId].reject = reject;
            
            setTimeout(() => {
                if (this.chrequest[reqId]) {
                    delete this.chrequest[reqId];
                    reject(`Timeout of ${timeout}ms expired`);
                }
            }, timeout);
        });

        await this.send({
            cmd: 'AUTH',
            topic: username,
            requestId: reqId,
            payload: password
        });

        return promise;
    }

    async send(data) {

        if(typeof data.payload === 'object') {
            data.payload = JSON.stringify(data.payload);
        }

        return new Promise((resolve, reject) => {
            const str = JSON.stringify(data) + '\n';
            this.conn.write(str, (err) => {
                if (err) reject(err);
                else resolve();
            });
        });
    }

    service(topic, fn) {
        this.services[topic] = fn;
        this.send({
            cmd: 'SER',
            topic: topic,
            payload: ''
        }).catch(err => this.emit('error', err));
    }

    subscribe(topic, cb) {
        if (!this.subs[topic]) {
            this.subs[topic] = [];
        }
        this.subs[topic].push(cb);
        
        this.send({
            cmd: 'SUB',
            topic: topic,
            payload: ''
        }).catch(err => this.emit('error', err));
    }

    publish(topic, payload) {
        this.send({
            cmd: 'PUB',
            topic: topic,
            payload: payload
        }).catch(err => this.emit('error', err));
    }

    async request(topic, payload, timeout = 1000) {
        const reqId = this.generateUUID();
        this.chrequest[reqId] = { resolve: null, reject: null };
        
        const promise = new Promise((resolve, reject) => {
            this.chrequest[reqId].resolve = resolve;
            this.chrequest[reqId].reject = reject;
            
            setTimeout(() => {
                if (this.chrequest[reqId]) {
                    delete this.chrequest[reqId];
                    reject(`Timeout of ${timeout}ms expired`);
                }
            }, timeout);
        });

        await this.send({
            cmd: 'REQ',
            topic: topic,
            requestId: reqId,
            payload: payload
        });

        return promise;
    }

    async ping() {
        const reqId = this.generateUUID();
        this.chs[reqId] = { resolve: null, reject: null };
        
        const promise = new Promise((resolve, reject) => {
            this.chs[reqId].resolve = resolve;
            this.chs[reqId].reject = reject;
            
            setTimeout(() => {
                if (this.chs[reqId]) {
                    delete this.chs[reqId];
                    reject('Timeout of 1s expired');
                }
            }, 1000);
        });

        await this.send({
            cmd: 'PING',
            topic: '',
            requestId: reqId,
            payload: ''
        });

        return promise;
    }

    script(name) {
        return new ScriptJS(this, name);
    }

    kv(bucket) {
        return new KV(this, bucket);
    }

    async dbCreateCollection(name, indexName) {
        try {
            await this.request('DB_CC', name, indexName, 2000);
            return true;
        } catch (err) {
            throw err;
        }
    }

    async dbDeleteCollection(name) {
        try {
            await this.request('DB_CD', name, '', 2000);
            return true;
        } catch (err) {
            throw err;
        }
    }

    dbCollection(name) {
        return new DbCollection(this, name);
    }

    onData(data) {
        this.buffer += data.toString();
        
        let newlineIndex;
        while ((newlineIndex = this.buffer.indexOf('\n')) !== -1) {
            const message = this.buffer.substring(0, newlineIndex).trim();
            this.buffer = this.buffer.substring(newlineIndex + 1);
            
            if (message) {
                try {
                    const parsed = JSON.parse(message);
                    this.handleMessage(parsed);
                } catch (err) {
             
                    this.emit('error',`Failed to parse message: ${err.message}\nMessage: ${message}`);
                }
            }
        }
    }

    handleMessage(data) {
        switch (data.cmd) {
            case 'CNN':
                this.id = data.payload;
                if (this.chrequest[data.requestId]) {
                    this.chrequest[data.requestId].resolve(data.payload);
                    delete this.chrequest[data.requestId];
                }
                break;
                
            case 'ER_AUH':
                this.id = data.payload;
                if (this.chrequest[data.requestId]) {
                    this.chrequest[data.requestId].reject(data.error);
                    delete this.chrequest[data.requestId];
                }
                this.stop();
                break;
                
            case 'S_ADD':
            case 'S_DEL':
            case 'S_ENV':
            case 'S_JS':
            case 'S_RUN':
            case 'S_STOP':
            case 'S_APP':
            case 'RES':
            case 'SET':
            case 'GET':
            case 'DEL':
            case 'BDEL':
            case 'BADD':
            case 'BFK':
            case 'BFV':
            case 'DB_CC':
            case 'DB_CD':
            case 'DB_CI':
            case 'DB_CG':
            case 'DB_CR':
            case 'DB_CF':
            case 'DB_CU':
            case 'DB_CL':
                if (this.chrequest[data.requestId]) {
                    if (data.error) {
                        this.chrequest[data.requestId].reject(data.error);
                    } else {
                        let result = data.payload;
                        try {
                            result=JSON.parse(result)
                        } catch (error) {
                           result = data.payload;
                        }
                        this.chrequest[data.requestId].resolve(result);
                    }
                    delete this.chrequest[data.requestId];
                }
                break;
                
            case 'PONG':
                if (this.chs[data.requestId]) {
                    this.chs[data.requestId].resolve(data.payload);
                    delete this.chs[data.requestId];
                }
                break;
                
            case 'REQ':
                if (this.services[data.topic]) {
                         let result = data.payload;
                        try {
                            result=JSON.parse(result)
                             data.payload=result
                        } catch (error) {
               
                        }
                    this.services[data.topic](data, (err, payload) => {
                        this.send({
                            cmd: 'RES',
                            requestId: data.requestId,
                            replayId: data.replayId,
                            topic: data.topic,
                            error: err,
                            payload: payload
                        }).catch(e => this.emit('error', e));
                    });
                }
                break;
                
            case 'PUB':
                let topic = data.topic;
                if (data.regtopic && data.regtopic.includes('.*.')) {
                    topic = data.regtopic;
                }
      
                        try {
                           data.payload=JSON.parse(data.payload)
                        } catch (error) {
                            
                        }
                if (this.subs[topic]) {
                    this.subs[topic].forEach(sub => sub(data));
                }
                break;
        }
    }

    stop() {
        return new Promise((resolve) => {
            this.conn.end(() => resolve());
        });
    }
}

class ScriptJS {
    constructor(mq, name) {
        this.mq = mq;
        this.name = name;
    }

    async send(type, key = '', value = '') {
        let topic = '';
        if (this.name) {
            topic = this.name;
            if (key) {
                topic = `${this.name}:${key}`;
            }
        } else {
            topic = key;
        }

        return this.mq.request(type, topic, value, 2000);
    }

    async setEnv(env) {
        return this.send('S_ENV', '', JSON.stringify(env));
    }

    async setApp(app) {
        return this.send('S_APP', '', JSON.stringify(app));
    }

    async setScript(scripts) {
        return this.send('S_JS', '', JSON.stringify(scripts));
    }

    async run() {
        return this.send('S_RUN');
    }

    async stop() {
        return this.send('S_STOP');
    }

    async register(input) {
        input.id = this.name;
        return this.send('S_ADD', '', JSON.stringify(input));
    }

    async remove() {
        return this.send('S_DEL');
    }
}

class KV {
    constructor(mq, bucket) {
        this.mq = mq;
        this.bucket = bucket;
    }

    async send(type, key = '', value = '') {
        let topic = '';
        if (this.bucket) {
            topic = this.bucket;
            if (key) {
                topic = `${this.bucket}:${key}`;
            }
        } else {
            topic = key;
        }

        return this.mq.request(type, topic, value, 2000);
    }

    async set(key, value) {
        return this.send('SET', key, value);
    }

    async get(key) {
        return this.send('GET', key);
    }

    async del(key) {
        return this.send('DEL', key);
    }

    async filterByKey(filter) {
        return this.send('BFK', '', filter);
    }

    async filterByValue(filter) {
        return this.send('BFV', '', filter);
    }

    async createBucket() {
        return this.send('BADD');
    }

    async deleteBucket() {
        return this.send('BDEL');
    }
}

class DbCollection {
    constructor(mq, name) {
        this.mq = mq;
        this.name = name;
    }

    async send(type, data = '') {
        return this.mq.request(type, this.name, data, 2000);
    }

    async insert(doc) {
        return this.send('DB_CI', JSON.stringify(doc));
    }

    async findOne(id) {
        const result = await this.send('DB_CG', id);
        return JSON.parse(result);
    }

    async find(doc) {
        const result = await this.send('DB_CF', JSON.stringify(doc));
        return JSON.parse(result);
    }

    async findAll() {
        const result = await this.send('DB_CL');
        return JSON.parse(result);
    }

    async update(id, doc) {
        doc._id = id;
        return this.send('DB_CU', JSON.stringify(doc));
    }

    async delete(id) {
        return this.send('DB_CR', id);
    }
}

module.exports = { MQ };