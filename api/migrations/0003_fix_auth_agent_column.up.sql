-- write your up migration here --
ALTER TABLE `authentication` MODIFY agent VARCHAR(500) NOT NULL;
