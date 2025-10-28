import { MongoClient } from 'mongodb';

async function addMetricsToAtlas(metricsDocs) {
    const uri =  process.env.ATLAS_CONNECTION_STRING;
    const client = new MongoClient(uri);
    try {
        await client.connect();
        const database = client.db("github_metrics");

        for (const doc of metricsDocs) {
            const collName = doc.owner + "_" + doc.repo;
            const coll = database.collection(collName);
            const result = await coll.insertOne(doc);
            console.log(`A document was inserted into ${collName} with the _id: ${result.insertedId}`);
        }
    } finally {
        await client.close();
    }
}

export {
    addMetricsToAtlas,
}