# PyMongo example 1
from pymongo import MongoClient

client = MongoClient('mongodb://localhost:27017')
db = client.test_database
collection = db.test_collection
result = collection.insert_one({'name': 'Alice', 'age': 30})
print(result.inserted_id)

