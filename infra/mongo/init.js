// Sitemon — Mongo bootstrap. Runs once on first container start.
// Creates the app DB, collections, and indexes for check history.
db = db.getSiblingDB('sitemon');

db.createCollection('checks');
db.checks.createIndex({ url: 1, timestamp: -1 });
db.checks.createIndex({ timestamp: -1 });

db.createCollection('requests');
db.requests.createIndex({ url: 1, timestamp: -1 });

print('sitemon: mongo initialized (checks, requests + indexes)');
