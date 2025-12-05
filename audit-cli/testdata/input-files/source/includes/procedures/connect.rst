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

