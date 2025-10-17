==========================
Nested Code Block Test
==========================

This file tests code-block directives that are nested inside other directives.

Test 1: Code Block Inside Procedure Step
==========================================

.. procedure::
   :style: normal

   .. step:: Create a database connection

      First, establish a connection to the database:

      .. code-block:: javascript
         :copyable: true

         const { MongoClient } = require('mongodb');
         const client = new MongoClient('mongodb://localhost:27017');
         await client.connect();

   .. step:: Insert a document

      Next, insert a document into the collection:

      .. code-block:: javascript
         :copyable: true

         const db = client.db('myDatabase');
         const collection = db.collection('myCollection');
         const result = await collection.insertOne({ name: 'Alice', age: 30 });
         console.log('Inserted document:', result.insertedId);

   .. step:: Query the document

      Finally, query the document you just inserted:

      .. code-block:: javascript

         const doc = await collection.findOne({ name: 'Alice' });
         console.log('Found document:', doc);

Test 2: Code Block Inside Note Directive
==========================================

.. note::

   When using transactions, you must use a session:

   .. code-block:: python
      :emphasize-lines: 2,3

      client = MongoClient('mongodb://localhost:27017')
      session = client.start_session()
      with session.start_transaction():
          collection.insert_one({'x': 1}, session=session)
          collection.update_one({'x': 1}, {'$set': {'y': 2}}, session=session)

Test 3: Code Block Inside Important Directive
===============================================

.. important::

   Always validate user input before processing:

   .. code-block:: go

      func validateInput(input string) error {
          if len(input) == 0 {
              return errors.New("input cannot be empty")
          }
          if len(input) > 100 {
              return errors.New("input too long")
          }
          return nil
      }

Test 4: Deeply Nested Code Block
==================================

.. container:: example

   .. admonition:: Example: Multi-step Process

      This example shows a multi-step process:

      .. procedure::

         .. step:: Initialize the system

            .. code-block:: typescript

               interface Config {
                   host: string;
                   port: number;
               }

               const config: Config = {
                   host: 'localhost',
                   port: 27017
               };

         .. step:: Connect to the database

            .. code-block:: typescript

               import { MongoClient } from 'mongodb';

               const client = new MongoClient(`mongodb://${config.host}:${config.port}`);
               await client.connect();
               console.log('Connected successfully');

Test 5: Code Block Inside Warning
===================================

.. warning::

   Do not use this pattern in production:

   .. code-block:: sh

      # This is insecure!
      chmod 777 /var/lib/mongodb
      chown nobody:nobody /var/lib/mongodb

Test 6: Multiple Code Blocks in Same Parent
=============================================

.. tip::

   You can use either syntax for connecting:

   **Option 1: Connection String**

   .. code-block:: ruby

      require 'mongo'
      client = Mongo::Client.new('mongodb://localhost:27017/mydb')

   **Option 2: Hash Options**

   .. code-block:: ruby

      require 'mongo'
      client = Mongo::Client.new(['localhost:27017'], database: 'mydb')

Test 7: Code Block with No Language Inside Directive
======================================================

.. note::

   Here's a sample configuration file:

   .. code-block::

      {
        "database": {
          "host": "localhost",
          "port": 27017
        },
        "logging": {
          "level": "info"
        }
      }

