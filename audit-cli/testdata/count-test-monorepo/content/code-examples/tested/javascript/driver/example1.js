// Node.js Driver example 1
const { MongoClient } = require('mongodb');

async function main() {
  const client = new MongoClient('mongodb://localhost:27017');
  await client.connect();
  console.log('Connected successfully');
  await client.close();
}

main();

