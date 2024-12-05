-- Active: 1731493020520@@127.0.0.1@5432@xm
CREATE TYPE ORG_TYPE AS ENUM (
    'Corporations', 
    'NonProfit', 
    'Cooperative',
    'Sole Proprietorship'
);


CREATE TABLE IF NOT EXISTS "company" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(15) UNIQUE NOT NULL,
    "description" VARCHAR(3000),
    "amount_employees" INT NOT NULL,
    "registered" BOOLEAN NOT NULL,
    "type" ORG_TYPE NOT NULL
);

CREATE UNIQUE INDEX company_id_idx ON "company"("id");
CREATE UNIQUE INDEX name ON "company"("name");    


CREATE TABLE IF NOT EXISTS "users" (
    "id" UUID PRIMARY KEY,
    "email" VARCHAR(50) UNIQUE NOT NULL,
    "enc_password" VARCHAR(32) NOT NULL
);

CREATE UNIQUE INDEX users_id_idx ON "users"("id");
CREATE UNIQUE INDEX email ON "users"("email");    


CREATE TYPE EVENT_TYPE AS ENUM (
    'create_company', 
    'update_company', 
    'delete_company'
);

CREATE TABLE IF NOT EXISTS "events" (
    "id" UUID PRIMARY KEY,
    "type" EVENT_TYPE NOT NULL,
    "timestamp" TIMESTAMP NOT NULL,
    "entity_id" UUID NOT NULL
);

CREATE UNIQUE INDEX events_id_idx ON "events"("id");
CREATE INDEX entity_id ON "events"("entity_id");