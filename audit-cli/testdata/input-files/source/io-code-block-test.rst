==========================
IO Code Block Test
==========================

This file tests io-code-block directives with input and output sub-directives.

Test 1: Inline Input and Output
=================================

.. io-code-block::
   :copyable: true

   .. input::
      :language: javascript

      db.restaurants.aggregate( [ { $match: { category: "cafe" } } ] )

   .. output::
      :language: javascript

      [
         { _id: 1, category: 'café', status: 'Open' },
         { _id: 2, category: 'cafe', status: 'open' },
         { _id: 3, category: 'cafE', status: 'open' }
      ]

Test 2: File-based Input and Output
=====================================

.. io-code-block::

   .. input:: /code-examples/example.js
      :language: javascript

   .. output:: /code-examples/example-output.txt
      :language: text

Test 3: Python Example with Inline Code
=========================================

.. io-code-block::

   .. input::
      :language: python

      from pymongo import MongoClient
      client = MongoClient('mongodb://localhost:27017')
      db = client.test_database
      collection = db.test_collection
      result = collection.insert_one({'name': 'Alice', 'age': 30})
      print(result.inserted_id)

   .. output::
      :language: python

      ObjectId('507f1f77bcf86cd799439011')

Test 4: Shell Command Example
===============================

.. io-code-block::
   :copyable: true

   .. input::
      :language: sh

      mongosh --eval "db.users.find({age: {$gt: 25}})"

   .. output::
      :language: json

      [
        { "_id": 1, "name": "Alice", "age": 30 },
        { "_id": 2, "name": "Bob", "age": 35 }
      ]

Test 5: TypeScript Example
============================

.. io-code-block::

   .. input::
      :language: ts

      import { MongoClient } from 'mongodb';
      
      const client = new MongoClient('mongodb://localhost:27017');
      await client.connect();
      const db = client.db('mydb');
      const result = await db.collection('users').findOne({ name: 'Alice' });
      console.log(result);

   .. output::
      :language: json

      { "_id": 1, "name": "Alice", "age": 30, "email": "alice@example.com" }

Test 6: Nested Inside Procedure Step
======================================

.. procedure::

   .. step:: Query the database

      Run the following query:

      .. io-code-block::
         :copyable: true

         .. input::
            :language: javascript

            db.inventory.find({ status: "A" })

         .. output::
            :language: javascript

            [
               { _id: 1, item: "journal", status: "A" },
               { _id: 2, item: "notebook", status: "A" }
            ]

Test 7: Input Only (No Output)
================================

.. io-code-block::

   .. input::
      :language: go

      package main
      
      import (
          "context"
          "go.mongodb.org/mongo-driver/mongo"
          "go.mongodb.org/mongo-driver/mongo/options"
      )
      
      func main() {
          client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
          if err != nil {
              panic(err)
          }
          defer client.Disconnect(context.TODO())
      }

