import { MongoClient } from 'mongodb';

const client = new MongoClient('mongodb://localhost:27017');
await client.connect();
const db = client.db('mydb');
const result = await db.collection('users').findOne({ name: 'Alice' });
console.log(result);