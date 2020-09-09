-- Description: Add support for CU holidays
-- Up migration
CREATE TABLE `cornell_holidays` (
	`id` int NOT NULL AUTO_INCREMENT,
	`startDate` date,
	`endDate` date,
	`name` text,
	`hasClasses` tinyint(1),
	PRIMARY KEY (id)
);