const doc = await collection.findOne({ name: 'Alice' });
console.log('Found document:', doc);