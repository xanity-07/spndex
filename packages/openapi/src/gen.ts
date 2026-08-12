import fs from 'node:fs';

import { OpenAPI } from './index.js';

const replaceCustumFileTypesToOpenAPICompatible = (jsonString: string): string => {
    const searchPattern =
        /{"type":"object","properties":{"type":{"type":"string","enum":\["file"\]}},\s*"required":\["type"\]}/g;
    const replacement = `{"type":"string","format":"binary"}`;

    return jsonString.replace(searchPattern, replacement);
};

const formattedDocs = JSON.parse(
    replaceCustumFileTypesToOpenAPICompatible(JSON.stringify(OpenAPI))
);

const filePaths = ['./openapi.json', '../../apps/backend/static/openapi.json'];

filePaths.forEach((filePath) => {
    fs.writeFile(filePath, JSON.stringify(formattedDocs, null, 2), (err) => {
        if (err) {
            console.error(`Error writing to ${filePath}:`, err);
        }
    });
});
