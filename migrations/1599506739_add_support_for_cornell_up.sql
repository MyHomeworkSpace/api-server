-- Description: add support for cornell
-- Up migration
CREATE TABLE `cornell_courses` (
	`id` int NOT NULL AUTO_INCREMENT,
	`userId` int,
	`subject` varchar(7),
	`catalogNum` int,
	`title` text,
	`units` int,
	`rosterId` int,
	PRIMARY KEY (id)
);
CREATE TABLE `cornell_events` (
	`id` int NOT NULL AUTO_INCREMENT,
	`title` text,
	`userId` int,
	`subject` varchar(7),
	`catalogNum` int,
	`classNum` int,
	`component` varchar(5),
	`componentLong` text,
	`section` varchar(5),
	`campus` varchar(10),
	`campusLong` text,
	`location` varchar(5),
	`locationLong` text,
	`startDate` date,
	`endDate` date,
	`startTime` int,
	`endTime` int,
	`monday` tinyint(1),
	`tuesday` tinyint(1),
	`wednesday` tinyint(1),
	`thursday` tinyint(1),
	`friday` tinyint(1),
	`saturday` tinyint(1),
	`sunday` tinyint(1),
	`facility` text,
	`facilityLong` text,
	`building` text,
	PRIMARY KEY (id)
);