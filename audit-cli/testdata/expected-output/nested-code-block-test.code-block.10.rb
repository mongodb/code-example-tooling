require 'mongo'
client = Mongo::Client.new(['localhost:27017'], database: 'mydb')