-- Description: Add support for Columbia University
-- Up migration

CREATE TABLE `columbia_classes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `department` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `number` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `section` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `instructorName` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `instructorEmail` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `userID` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `columbia_holidays` (
  `id` int NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `text` text COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `columbia_meetings` (
  `id` int NOT NULL AUTO_INCREMENT,
  `department` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `number` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `section` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `building` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `room` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `dow` int NOT NULL,
  `start` int NOT NULL,
  `end` int NOT NULL,
  `beginDate` date NOT NULL,
  `endDate` date NOT NULL,
  `userID` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;