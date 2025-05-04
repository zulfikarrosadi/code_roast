-- write your down migration here --
ALTER TABLE authentication MODIFY agent VARCHAR(100) NOT NULL;
