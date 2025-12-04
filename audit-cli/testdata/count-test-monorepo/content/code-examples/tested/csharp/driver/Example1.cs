// C#/.NET Driver example 1
using MongoDB.Driver;

class Example1
{
    static void Main()
    {
        var client = new MongoClient("mongodb://localhost:27017");
        var database = client.GetDatabase("test");
        Console.WriteLine("Connected successfully");
    }
}

