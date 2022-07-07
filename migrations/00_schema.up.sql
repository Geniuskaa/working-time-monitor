CREATE TABLE Employees (
    id serial primary key,
    name text not null
);

CREATE TABLE users
(
    id        SERIAL PRIMARY KEY,
    username     TEXT      NOT NULL UNIQUE,
    display_name  TEXT      NOT NULL,
    empl_id INTEGER NOT NULL REFERENCES  Employees(id),
    email  TEXT      NOT NULL unique,
    phone TEXT not null unique,
    birthday  DATE      NOT NULL,
    skills TEXT NOT NULL ,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE projects
(
    id serial primary key,
    name text not null
);

CREATE TABLE user_projects
(
    id serial primary key,
    user_id integer references users(id),
    project_id integer references projects(id)
);





