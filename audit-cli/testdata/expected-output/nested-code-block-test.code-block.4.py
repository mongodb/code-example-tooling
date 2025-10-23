client = MongoClient('mongodb://localhost:27017')
session = client.start_session()
with session.start_transaction():
    collection.insert_one({'x': 1}, session=session)
    collection.update_one({'x': 1}, {'$set': {'y': 2}}, session=session)