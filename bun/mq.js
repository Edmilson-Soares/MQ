const net = require('net');
const crypto = require('node:crypto');

class MQData {
    constructor({ cmd, topic, payload, requestId, replayId, regtopic, error, fromId }) {
        this.cmd = cmd;
        this.topic = topic;
        this.payload = payload;
        this.requestId = requestId;
        this.replayId = replayId;
        this.regtopic = regtopic;
        this.error = error;
        this.fromId = fromId;
    }
}

class MQResponse {
    constructor({ payload, error }) {
        this.payload = payload;
        this.error = error;
    }
}

function jsonToStructResponse(data) {
    try {
        return new MQResponse(JSON.parse(data));
    } catch (err) {
        throw err;
    }
}

function structToJSONResponse(res) {
    try {
        return JSON.stringify(res);
    } catch (err) {
        throw err;
    }
}

function jsonToStruct(data) {
    try {
        return new MQData(JSON.parse(data));
    } catch (err) {
        throw err;
    }
}

function structToJSON(mq) {
    try {
        return JSON.stringify(mq);
    } catch (err) {
        throw err;
    }
}

class MQ {
    constructor() {
        this.ID = '';
        this.conn = null;
        this.subs = new Map();
        this.chs = new Map();
        this.chrequest = new Map();
        this.services = new Map();
    }

    static async connect(url) {
        const info = MQ.parseMQURL(url);
        const mq = new MQ();
        
        return new Promise((resolve, reject) => {
            const conn = net.createConnection({
                host: info.host,
                port: info.port
            }, () => {
                mq.conn = conn;
                mq.send(new MQData({
                    cmd: "AUTH",
                    topic: "",
                    payload: info.user + ":" + info.pass
                }));
                mq.on();
                resolve(mq);
            });
            
            conn.on('error', (err) => {
                reject(err);
            });
        });
    }

    static parseMQURL(url) {
        // Basic URL parsing - you might want to enhance this
        const parts = url.split('://')[1].split('@');
        const [user, pass] = parts[0].split(':');
        const [host, port] = parts[1].split(':');
        return { user, pass, host, port };
    }

    send(data) {
        return new Promise((resolve, reject) => {
            try {
                const str = structToJSON(data);
                this.conn.write(str + '\n', (err) => {
                    if (err) reject(err);
                    else resolve();
                });
            } catch (err) {
                reject(err);
            }
        });
    }

    service(topic, fn) {
        this.services.set(topic, fn);
        this.send(new MQData({
            cmd: "SER",
            topic: topic,
            payload: ""
        })).catch(console.error);
    }

    sub(topic, cb) {
        if (!this.subs.has(topic)) {
            this.subs.set(topic, []);
        }
        this.subs.get(topic).push(cb);
        this.send(new MQData({
            cmd: "SUB",
            topic: topic,
            payload: ""
        })).catch(console.error);
    }

    pub(topic, payload) {
        this.send(new MQData({
            cmd: "PUB",
            topic: topic,
            payload: payload
        })).catch(console.error);
    }

    async req(topic, payload, timeout = 1000) {
        const reqId = crypto.randomUUID();
        this.chrequest.set(reqId, {
            response: null,
            resolve: null,
            reject: null
        });

        await this.send(new MQData({
            cmd: "REQ",
            topic: topic,
            requestId: reqId,
            payload: payload
        }));

        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.chrequest.delete(reqId);
                reject(new Error(`timeout of ${timeout}ms expired for topic ${topic}`));
            }, timeout);

            this.chrequest.get(reqId).resolve = (res) => {
                clearTimeout(timer);
                this.chrequest.delete(reqId);
                if (res.error) {
                    reject(new Error("Error: " + res.error));
                } else {
                    resolve(res.payload);
                }
            };

            this.chrequest.get(reqId).reject = (err) => {
                clearTimeout(timer);
                this.chrequest.delete(reqId);
                reject(err);
            };
        });
    }

    async ping() {
        const reqId = crypto.randomUUID();
        this.chs.set(reqId, {
            response: null,
            resolve: null,
            reject: null
        });

        await this.send(new MQData({
            cmd: "PING",
            topic: "",
            requestId: reqId,
            payload: ""
        }));

        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.chs.delete(reqId);
                reject(new Error(`timeout of 1000ms expired for ping`));
            }, 1000);

            this.chs.get(reqId).resolve = (res) => {
                clearTimeout(timer);
                this.chs.delete(reqId);
                resolve(res);
            };

            this.chs.get(reqId).reject = (err) => {
                clearTimeout(timer);
                this.chs.delete(reqId);
                reject(err);
            };
        });
    }

    async get(key) {
        return this.req("GET", key);
    }

    async set(key, value) {
        return this.req("SET", value, 1000, key);
    }

    async del(key) {
        return this.req("DEL", "", 1000, key);
    }

    async filterKey(filter) {
        return this.req("BFK", "", 1000, filter);
    }

    async filterValue(filter) {
        return this.req("BFV", "", 1000, filter);
    }

    async bFilterKey(bucket, filter) {
        return this.req("BFK", "", 1000, bucket + ":" + filter);
    }

    async bFilterValue(bucket, filter) {
        return this.req("BFV", "", 1000, bucket + ":" + filter);
    }

    async bGet(bucket, key) {
        return this.req("GET", "", 1000, bucket + ":" + key);
    }

    async bSet(bucket, key, value) {
        return this.req("SET", value, 1000, bucket + ":" + key);
    }

    async bDel(bucket, key) {
        return this.req("DEL", "", 1000, bucket + ":" + key);
    }

    async deleteBucket(name) {
        return this.req("BDEL", "", 1000, name);
    }

    async createBucket(name) {
        return this.req("BADD", "", 1000, name);
    }

    on() {
        const reader = require('readline').createInterface({
            input: this.conn
        });

        reader.on('line', (str) => {
            try {
                const data = jsonToStruct(str);
                
                switch (data.cmd) {
                    case "CNN":
                        this.ID = data.payload;
                        console.log("connected");
                        break;
                    case "OK":
                        // Do nothing
                        break;
                    case "ER_AUH":
                        this.stop();
                        throw new Error(data.payload);
                    case "RES":
                    case "SET":
                    case "GET":
                    case "DEL":
                    case "BDEL":
                    case "BADD":
                    case "BFK":
                    case "BFV":
                        const ch = this.chrequest.get(data.requestId);
                        if (ch) {
                            ch.resolve(new MQResponse({
                                payload: data.payload,
                                error: data.error
                            }));
                        }
                        break;
                    case "PONG":
                        const pingCh = this.chs.get(data.requestId);
                        if (pingCh) {
                            pingCh.resolve(data.payload);
                        }
                        break;
                    case "REQ":
                        const service = this.services.get(data.topic);
                        if (service) {
                            service(data, (err, payload) => {
                                this.send(new MQData({
                                    cmd: "RES",
                                    requestId: data.requestId,
                                    replayId: data.replayId,
                                    topic: data.topic,
                                    error: err,
                                    payload: payload
                                })).catch(console.error);
                            });
                        }
                        break;
                    case "PUB":
                        let topic = data.topic;
                        if (data.regtopic && data.regtopic.includes('.*.')) {
                            topic = data.regtopic;
                        }
                        const subscribers = this.subs.get(topic) || [];
                        subscribers.forEach(sub => sub(data));
                        break;
                }
            } catch (err) {
                console.error(`Error processing message: ${err}`);
            }
        });

        reader.on('close', () => {
            console.log('Connection closed');
        });

        this.conn.on('error', (err) => {
            console.error(`Connection error: ${err}`);
        });

        this.conn.on('end', () => {
            console.log('Connection ended');
            reader.close();
        });
    }

    stop() {
        if (this.conn) {
            this.conn.end();
        }
    }
}

module.exports = {
    MQ,
    MQData,
    MQResponse
};