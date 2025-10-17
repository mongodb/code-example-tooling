Code Block Test
===============

This file tests various code-block directive scenarios.

JavaScript with Language
------------------------

.. code-block:: javascript

   const greeting = "Hello, World!";
   console.log(greeting);

Python with Options
-------------------

.. code-block:: python
   :copyable: false
   :emphasize-lines: 2,3

   def calculate_sum(a, b):
       result = a + b
       return result

JSON Array Example
------------------

.. code-block:: javascript
   :copyable: false
   :emphasize-lines: 12,13,25,26,31,32

   [
     {
       _id: ObjectId("620ad555394d47411658b5ef"),
       time: ISODate("2021-03-08T09:00:00.000Z"),
       price: 500,
       linearFillPrice: 500,
       locfPrice: 500
     },
     {
       _id: ObjectId("620ad555394d47411658b5f0"),
       time: ISODate("2021-03-08T10:00:00.000Z"),
       linearFillPrice: 507.5,
       locfPrice: 500
     },
     {
       _id: ObjectId("620ad555394d47411658b5f1"),
       time: ISODate("2021-03-08T11:00:00.000Z"),
       price: 515,
       linearFillPrice: 515,
       locfPrice: 515
     },
     {
       _id: ObjectId("620ad555394d47411658b5f2"),
       time: ISODate("2021-03-08T12:00:00.000Z"),
       linearFillPrice: 505,
       locfPrice: 515
     },
     {
       _id: ObjectId("620ad555394d47411658b5f3"),
       time: ISODate("2021-03-08T13:00:00.000Z"),
       linearFillPrice: 495,
       locfPrice: 515
     },
     {
       _id: ObjectId("620ad555394d47411658b5f4"),
       time: ISODate("2021-03-08T14:00:00.000Z"),
       price: 485,
       linearFillPrice: 485,
       locfPrice: 485
     }
   ]

Code Block with No Language
----------------------------

.. code-block::

   This is a code block with no language specified.
   It should still be extracted.

Shell Script
------------

.. code-block:: sh

   #!/bin/bash
   echo "Hello from shell"
   exit 0

TypeScript Normalization
------------------------

.. code-block:: ts

   interface User {
       name: string;
       age: number;
   }

C++ Normalization
-----------------

.. code-block:: c++

   #include <iostream>
   
   int main() {
       std::cout << "Hello" << std::endl;
       return 0;
   }

