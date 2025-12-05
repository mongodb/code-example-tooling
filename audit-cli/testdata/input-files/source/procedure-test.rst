============================
Procedure Testing Examples
============================

This file contains various procedure examples for testing the procedure parsing and extraction functionality.

Simple Procedure with Steps
============================

.. procedure::
   :style: normal

   .. step:: Create a database connection

      First, establish a connection to the database:

      .. code-block:: javascript

         const { MongoClient } = require('mongodb');
         const client = new MongoClient('mongodb://localhost:27017');
         await client.connect();

   .. step:: Insert a document

      Next, insert a document into the collection:

      .. code-block:: javascript

         const db = client.db('myDatabase');
         const collection = db.collection('myCollection');
         const result = await collection.insertOne({ name: 'Alice', age: 30 });

   .. step:: Close the connection

      Finally, close the connection:

      .. code-block:: javascript

         await client.close();

Procedure with Tabs
====================

.. procedure::

   .. step:: Connect to MongoDB

      Choose your preferred connection method:

      .. tabs::

         .. tab:: MongoDB Shell
            :tabid: shell

            Connect using the MongoDB Shell:

            .. code-block:: bash

               mongosh "mongodb://localhost:27017"

         .. tab:: Node.js Driver
            :tabid: nodejs

            Connect using the Node.js driver:

            .. code-block:: javascript

               const { MongoClient } = require('mongodb');
               const client = new MongoClient('mongodb://localhost:27017');
               await client.connect();

         .. tab:: Python Driver
            :tabid: python

            Connect using the Python driver:

            .. code-block:: python

               from pymongo import MongoClient
               client = MongoClient('mongodb://localhost:27017')

   .. step:: Verify the connection

      Verify that you're connected:

      .. tabs::

         .. tab:: MongoDB Shell
            :tabid: shell

            .. code-block:: bash

               db.runCommand({ ping: 1 })

         .. tab:: Node.js Driver
            :tabid: nodejs

            .. code-block:: javascript

               await client.db('admin').command({ ping: 1 });

         .. tab:: Python Driver
            :tabid: python

            .. code-block:: python

               client.admin.command('ping')

Composable Tutorial Example
============================

.. composable-tutorial::
   :options: interface, language
   :defaults: driver, nodejs

   .. procedure::

      .. step:: Install dependencies

         .. selected-content::
            :selections: driver, nodejs

            Install the MongoDB Node.js driver:

            .. code-block:: bash

               npm install mongodb

         .. selected-content::
            :selections: driver, python

            Install the MongoDB Python driver:

            .. code-block:: bash

               pip install pymongo

         .. selected-content::
            :selections: atlas-cli, none

            Install the Atlas CLI:

            .. code-block:: bash

               brew install mongodb-atlas-cli

      .. step:: Connect to your cluster

         .. selected-content::
            :selections: driver, nodejs

            Create a connection using the Node.js driver:

            .. code-block:: javascript

               const { MongoClient } = require('mongodb');
               const uri = process.env.MONGODB_URI;
               const client = new MongoClient(uri);
               await client.connect();

         .. selected-content::
            :selections: driver, python

            Create a connection using the Python driver:

            .. code-block:: python

               from pymongo import MongoClient
               import os
               uri = os.environ['MONGODB_URI']
               client = MongoClient(uri)

         .. selected-content::
            :selections: atlas-cli, none

            Authenticate with the Atlas CLI:

            .. code-block:: bash

               atlas auth login

      .. step:: Perform an operation

         General content that applies to all selections.

         .. selected-content::
            :selections: driver, nodejs

            Insert a document using Node.js:

            .. code-block:: javascript

               const db = client.db('test');
               const result = await db.collection('users').insertOne({ name: 'Alice' });
               console.log('Inserted document:', result.insertedId);

         .. selected-content::
            :selections: driver, python

            Insert a document using Python:

            .. code-block:: python

               db = client.test
               result = db.users.insert_one({'name': 'Alice'})
               print('Inserted document:', result.inserted_id)

         .. selected-content::
            :selections: atlas-cli, none

            Create a cluster using the Atlas CLI:

            .. code-block:: bash

               atlas clusters create myCluster --provider AWS --region US_EAST_1

Ordered List Procedure
=======================

1. Create a new directory for your project:

   .. code-block:: bash

      mkdir my-mongodb-project
      cd my-mongodb-project

2. Initialize a new Node.js project:

   .. code-block:: bash

      npm init -y

3. Install the MongoDB driver:

   .. code-block:: bash

      npm install mongodb

4. Create a connection file:

   .. code-block:: javascript

      const { MongoClient } = require('mongodb');
      const uri = 'mongodb://localhost:27017';
      const client = new MongoClient(uri);

Procedure with Sub-steps
=========================

.. procedure::

   .. step:: Set up your environment

      a. Install Node.js from https://nodejs.org
      b. Install MongoDB from https://www.mongodb.com/try/download/community
      c. Verify installations:

         .. code-block:: bash

            node --version
            mongod --version

   .. step:: Create your project

      a. Create a new directory
      b. Initialize npm
      c. Install dependencies

      .. code-block:: bash

         mkdir my-app && cd my-app
         npm init -y
         npm install mongodb

   .. step:: Write your code

      Create an `index.js` file with the following content:

      .. code-block:: javascript

         const { MongoClient } = require('mongodb');
         
         async function main() {
           const client = new MongoClient('mongodb://localhost:27017');
           await client.connect();
           console.log('Connected to MongoDB');
           await client.close();
         }
         
         main().catch(console.error);

