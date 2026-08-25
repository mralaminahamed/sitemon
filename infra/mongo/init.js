// Sitemon — Mongo bootstrap. Runs once on first container start.
// Creates the app DB, collections, and indexes for check history.
db = db.getSiblingDB('sitemon');

db.createCollection('checks');
db.checks.createIndex({ url: 1, timestamp: -1 });
// TTL: expire check history after 30 days to cap unbounded growth.
db.checks.createIndex({ timestamp: 1 }, { expireAfterSeconds: 2592000 });

db.createCollection('alerts');
db.alerts.createIndex({ url: 1, timestamp: -1 });
// TTL: keep alert history 90 days.
db.alerts.createIndex({ timestamp: 1 }, { expireAfterSeconds: 7776000 });

db.createCollection('requests');
db.requests.createIndex({ url: 1, timestamp: -1 });

print('sitemon: mongo initialized (checks, alerts, requests + TTL indexes)');
