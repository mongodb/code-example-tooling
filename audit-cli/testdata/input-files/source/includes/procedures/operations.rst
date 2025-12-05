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

