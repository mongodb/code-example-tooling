Literalinclude Test
===================

This file tests various literalinclude directive scenarios.

Python with start-after and end-before
---------------------------------------

.. literalinclude:: /code-examples/example.py
   :language: python
   :start-after: start-hello
   :end-before: end-hello
   :dedent:

Go full file
------------

.. literalinclude:: /code-examples/example.go
   :language: go

JavaScript with start-after only
---------------------------------

.. literalinclude:: /code-examples/example.js
   :language: javascript
   :start-after: start-greet

PHP with end-before only
-------------------------

.. literalinclude:: /code-examples/example.php
   :language: php
   :end-before: end-init

Ruby with dedent
----------------

.. literalinclude:: /code-examples/example.rb
   :language: ruby
   :dedent:

TypeScript language normalization
----------------------------------

.. literalinclude:: /code-examples/example.ts
   :language: ts

C++ language normalization
---------------------------

.. literalinclude:: /code-examples/example.cpp
   :language: c++

