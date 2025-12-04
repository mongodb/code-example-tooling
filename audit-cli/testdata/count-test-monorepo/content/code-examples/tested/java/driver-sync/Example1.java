// Java Sync Driver example 1
import com.mongodb.client.MongoClient;
import com.mongodb.client.MongoClients;

public class Example1 {
    public static void main(String[] args) {
        MongoClient client = MongoClients.create("mongodb://localhost:27017");
        System.out.println("Connected successfully");
        client.close();
    }
}

