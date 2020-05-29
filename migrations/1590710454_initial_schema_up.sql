-- Description: Initial schema
-- Up migration

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

SET NAMES utf8mb4;


CREATE TABLE `applications` (
  `id` int NOT NULL AUTO_INCREMENT,
  `userId` int DEFAULT NULL,
  `name` text COLLATE utf8mb4_unicode_ci,
  `authorName` text COLLATE utf8mb4_unicode_ci,
  `clientId` text COLLATE utf8mb4_unicode_ci,
  `callbackUrl` text COLLATE utf8mb4_unicode_ci,
  `cors` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `application_authorizations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `userId` int DEFAULT NULL,
  `applicationId` int DEFAULT NULL,
  `token` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `calendar_events` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` text COLLATE utf8mb4_unicode_ci,
  `start` int DEFAULT NULL,
  `end` int DEFAULT NULL,
  `desc` text COLLATE utf8mb4_unicode_ci,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `calendar_event_changes` (
  `eventID` varchar(180) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cancel` tinyint(1) NOT NULL,
  `userID` int NOT NULL,
  PRIMARY KEY (`eventID`,`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `calendar_event_rules` (
  `id` int NOT NULL AUTO_INCREMENT,
  `eventId` int DEFAULT NULL,
  `frequency` tinyint(1) DEFAULT NULL,
  `interval` int DEFAULT NULL,
  `byDay` varchar(45) COLLATE utf8mb4_unicode_ci NOT NULL,
  `byMonthDay` int DEFAULT NULL,
  `byMonth` int DEFAULT NULL,
  `until` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `calendar_external` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `url` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `lastUpdated` int NOT NULL,
  `enabled` tinyint(1) NOT NULL,
  `hidden` tinyint(1) NOT NULL,
  `userID` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `calendar_external_events` (
  `uid` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `start` int NOT NULL,
  `end` int NOT NULL,
  `calendarID` int NOT NULL,
  KEY `calendarID` (`calendarID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `calendar_hwevents` (
  `id` int NOT NULL AUTO_INCREMENT,
  `homeworkId` int DEFAULT NULL,
  `start` int DEFAULT NULL,
  `end` int DEFAULT NULL,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `classes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `teacher` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `color` varchar(6) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sortIndex` int NOT NULL,
  `userId` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `dalton_announcements` (
  `id` int NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `text` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `grade` int NOT NULL,
  `type` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `dalton_classes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `termId` int DEFAULT NULL,
  `ownerId` int DEFAULT NULL,
  `sectionId` int DEFAULT NULL,
  `name` text COLLATE utf8mb4_unicode_ci,
  `ownerName` text COLLATE utf8mb4_unicode_ci,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `dalton_fridays` (
  `id` int NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `index` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `dalton_periods` (
  `id` int NOT NULL AUTO_INCREMENT,
  `classId` int DEFAULT NULL,
  `dayNumber` int DEFAULT NULL,
  `block` text COLLATE utf8mb4_unicode_ci,
  `buildingName` text COLLATE utf8mb4_unicode_ci,
  `roomNumber` text COLLATE utf8mb4_unicode_ci,
  `start` int DEFAULT NULL,
  `end` int DEFAULT NULL,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `dalton_terms` (
  `id` int NOT NULL AUTO_INCREMENT,
  `termId` int DEFAULT NULL,
  `name` text COLLATE utf8mb4_unicode_ci,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `feedback` (
  `id` int NOT NULL AUTO_INCREMENT,
  `userId` int DEFAULT NULL,
  `type` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `text` text COLLATE utf8mb4_unicode_ci,
  `screenshot` longblob,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `homework` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` text COLLATE utf8mb4_unicode_ci,
  `due` date DEFAULT NULL,
  `desc` text COLLATE utf8mb4_unicode_ci,
  `complete` tinyint(1) DEFAULT NULL,
  `classId` int DEFAULT NULL,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `internal_tasks` (
  `taskID` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `lastCompletion` date NOT NULL,
  PRIMARY KEY (`taskID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `mit_classes` (
  `subjectID` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sectionID` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `units` int NOT NULL,
  `sections` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `userID` int NOT NULL,
  PRIMARY KEY (`subjectID`,`sectionID`,`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `mit_holidays` (
  `id` int NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `text` text COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `mit_listings` (
  `id` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `shortTitle` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `offeredFall` tinyint(1) NOT NULL,
  `offeredIAP` tinyint(1) NOT NULL,
  `offeredSpring` tinyint(1) NOT NULL,
  `fallInstructors` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `springInstructors` text COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `mit_offerings` (
  `id` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `section` varchar(6) COLLATE utf8mb4_unicode_ci NOT NULL,
  `term` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL,
  `time` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `place` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `facultyID` varchar(9) COLLATE utf8mb4_unicode_ci NOT NULL,
  `facultyName` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `isFake` tinyint(1) NOT NULL,
  `isMaster` tinyint(1) NOT NULL,
  `isDesign` tinyint(1) NOT NULL,
  `isLab` tinyint(1) NOT NULL,
  `isLecture` tinyint(1) NOT NULL,
  `isRecitation` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`,`section`,`term`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `notifications` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `content` text COLLATE utf8mb4_unicode_ci,
  `expiry` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `prefixes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `words` text COLLATE utf8mb4_unicode_ci,
  `color` varchar(6) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `background` varchar(6) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `isTimedEvent` tinyint(1) DEFAULT NULL,
  `userId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `prefs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `userId` int DEFAULT NULL,
  `key` text COLLATE utf8mb4_unicode_ci,
  `value` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `schools` (
  `id` int NOT NULL AUTO_INCREMENT,
  `schoolId` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `enabled` tinyint(1) NOT NULL,
  `data` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `userId` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `tabs` (
  `id` int NOT NULL,
  `slug` text COLLATE utf8mb4_unicode_ci,
  `icon` text COLLATE utf8mb4_unicode_ci,
  `label` text COLLATE utf8mb4_unicode_ci,
  `target` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `tab_permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `userId` int DEFAULT NULL,
  `tabId` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `totp` (
  `userId` int NOT NULL,
  `secret` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `features` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '[]',
  `emailVerified` tinyint(1) NOT NULL,
  `level` tinyint(1) NOT NULL DEFAULT '0',
  `canFeedback` tinyint(1) NOT NULL DEFAULT '0',
  `canAnnouncements` tinyint(1) NOT NULL DEFAULT '0',
  `showMigrateMessage` tinyint(1) DEFAULT NULL,
  `createdAt` int NOT NULL,
  `lastLoginAt` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;