CREATE TABLE services_new (id TEXT NOT NULL PRIMARY KEY,
						   aosVersion INTEGER,															   
						   serviceProvider TEXT,
						   path TEXT,
						   unit TEXT,
						   uid INTEGER,
						   gid INTEGER,															   
						   state INTEGER,															   
						   startat TIMESTAMP,															   
						   alertRules TEXT,															   														   
						   vendorVersion TEXT,
						   description TEXT,
						   manifestDigest BLOB);

INSERT INTO  services_new (id, aosVersion, serviceProvider, path, unit, uid, gid,
        state, startat, alertRules, vendorVersion, description, manifestDigest)
SELECT id, aosVersion, serviceProvider, path, unit, uid, gid,
        state, startat, alertRules, vendorVersion, description, manifestDigest
FROM services;

DROP TABLE services;

ALTER TABLE services_new RENAME TO services;