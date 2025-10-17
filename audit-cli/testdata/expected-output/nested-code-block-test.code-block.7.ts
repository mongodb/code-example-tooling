import { MongoClient } from 'mongodb';

const client = new MongoClient(`mongodb://${config.host}:${config.port}`);
await client.connect();
console.log('Connected successfully');