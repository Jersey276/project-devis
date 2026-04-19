CREATE TABLE countries (
    id   SERIAL   PRIMARY KEY,
    code CHAR(2)  NOT NULL UNIQUE,
    name TEXT     NOT NULL
);
