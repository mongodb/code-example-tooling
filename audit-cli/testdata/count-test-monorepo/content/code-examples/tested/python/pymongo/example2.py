# PyMongo example 2
from pymongo import MongoClient

client = MongoClient('mongodb://localhost:27017')
result = client.admin.command('ping')
print(result)

