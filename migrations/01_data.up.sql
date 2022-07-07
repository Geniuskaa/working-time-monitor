INSERT INTO Employees(name)
VALUES ('Go-developer'),
('Ios-developer'),
('Java-developer'),
('Android-developer');

INSERT INTO users(username, display_name,empl_id, email, phone, birthday, skills, created_at, updated_at)
VALUES ('lurik', 'Vladislav', 2, 'vladik@mail.ru', '+79572286256','2000-02-22', 'Swift, SQLite', now(), now()),
       ('markov', 'Oleg', 1, 'oleja@mail.ru', '+79502286256','1990-07-09', 'Go, MongoDB', now(), now()),
       ('genesis', 'Ridvan', 3, 'genya@mail.ru', '+79572106256','2002-01-05', 'Java, PostgreSQL', now(), now()),
       ('satoshi', 'Satoshi', 4, 'satoshik@mail.ru', '+79570286756','2001-06-16', 'Kotlin, PostgreSQL', now(), now());


INSERT INTO projects (name) VALUES ('Халвёнок'), ('Совёнок'), ('Кутёнок');

INSERT INTO user_projects (user_id, project_id)
VALUES (1,2),(2,3),(3,1),(4,2);