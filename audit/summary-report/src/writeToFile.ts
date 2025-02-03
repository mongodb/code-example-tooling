import { promises as fs } from 'fs';
import * as path from 'path';
export async function writeJSONToFile(filePath: string, data: any): Promise<void> {
    try {
        // Convert the data object to a JSON string
        const jsonString = JSON.stringify(data, null, 2); // Pretty print with 2 spaces
        // Write the JSON string to the file
        await fs.writeFile(filePath, jsonString, 'utf8');
        console.log(`Successfully wrote to ${filePath}`);
    } catch (error) {
        console.error(`Error writing file ${filePath}:`, error);
    }
}
