

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