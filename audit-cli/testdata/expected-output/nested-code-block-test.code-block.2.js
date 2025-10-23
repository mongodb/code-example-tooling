const db = client.db('myDatabase');
const collection = db.collection('myCollection');
const result = await collection.insertOne({ name: 'Alice', age: 30 });
console.log('Inserted document:', result.insertedId);