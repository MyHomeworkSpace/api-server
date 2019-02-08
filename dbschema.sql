# ************************************************************
# Sequel Pro SQL dump
# Version 4541
#
# http://www.sequelpro.com/
# https://github.com/sequelpro/sequelpro
#
# Host: 127.0.0.1 (MySQL 5.7.18)
# Database: myhomeworkspace
# Generation Time: 2019-02-08 04:11:36 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table announcements
# ------------------------------------------------------------

CREATE TABLE `announcements` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `text` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `grade` int(2) NOT NULL,
  `type` int(2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table application_authorizations
# ------------------------------------------------------------

CREATE TABLE `application_authorizations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `applicationId` int(11) DEFAULT NULL,
  `token` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table applications
# ------------------------------------------------------------

CREATE TABLE `applications` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `name` mediumtext COLLATE utf8mb4_unicode_ci,
  `authorName` mediumtext COLLATE utf8mb4_unicode_ci,
  `clientId` mediumtext COLLATE utf8mb4_unicode_ci,
  `callbackUrl` mediumtext COLLATE utf8mb4_unicode_ci,
  `cors` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table calendar_classes
# ------------------------------------------------------------

CREATE TABLE `calendar_classes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `termId` int(11) DEFAULT NULL,
  `ownerId` int(11) DEFAULT NULL,
  `sectionId` int(11) DEFAULT NULL,
  `name` mediumtext COLLATE utf8mb4_unicode_ci,
  `ownerName` mediumtext COLLATE utf8mb4_unicode_ci,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table calendar_events
# ------------------------------------------------------------

CREATE TABLE `calendar_events` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` mediumtext COLLATE utf8mb4_unicode_ci,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `desc` mediumtext COLLATE utf8mb4_unicode_ci,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table calendar_hwevents
# ------------------------------------------------------------

CREATE TABLE `calendar_hwevents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `homeworkId` int(11) DEFAULT NULL,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table calendar_periods
# ------------------------------------------------------------

CREATE TABLE `calendar_periods` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `classId` int(11) DEFAULT NULL,
  `dayNumber` int(11) DEFAULT NULL,
  `block` mediumtext COLLATE utf8mb4_unicode_ci,
  `buildingName` mediumtext COLLATE utf8mb4_unicode_ci,
  `roomNumber` mediumtext COLLATE utf8mb4_unicode_ci,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table calendar_status
# ------------------------------------------------------------

CREATE TABLE `calendar_status` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table calendar_terms
# ------------------------------------------------------------

CREATE TABLE `calendar_terms` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `termId` int(11) DEFAULT NULL,
  `name` mediumtext COLLATE utf8mb4_unicode_ci,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table classes
# ------------------------------------------------------------

CREATE TABLE `classes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` mediumtext COLLATE utf8mb4_unicode_ci,
  `teacher` mediumtext COLLATE utf8mb4_unicode_ci,
  `color` varchar(12) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table feedback
# ------------------------------------------------------------

CREATE TABLE `feedback` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `type` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `text` mediumtext COLLATE utf8mb4_unicode_ci,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `screenshot` blob,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table fridays
# ------------------------------------------------------------

CREATE TABLE `fridays` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `index` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table homework
# ------------------------------------------------------------

CREATE TABLE `homework` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` mediumtext COLLATE utf8mb4_unicode_ci,
  `due` date DEFAULT NULL,
  `desc` mediumtext COLLATE utf8mb4_unicode_ci,
  `complete` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `classId` int(11) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table notifications
# ------------------------------------------------------------

CREATE TABLE `notifications` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `content` text COLLATE utf8mb4_unicode_ci,
  `expiry` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table prefixes
# ------------------------------------------------------------

CREATE TABLE `prefixes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `words` text COLLATE utf8mb4_unicode_ci,
  `color` varchar(6) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `background` varchar(6) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `isTimedEvent` tinyint(1) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table prefs
# ------------------------------------------------------------

CREATE TABLE `prefs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `key` mediumtext COLLATE utf8mb4_unicode_ci,
  `value` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table tab_permissions
# ------------------------------------------------------------

CREATE TABLE `tab_permissions` (
  `id` int(11) NOT NULL,
  `userId` int(11) DEFAULT NULL,
  `tabId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table tabs
# ------------------------------------------------------------

CREATE TABLE `tabs` (
  `id` int(11) NOT NULL,
  `slug` mediumtext COLLATE utf8mb4_unicode_ci,
  `icon` mediumtext COLLATE utf8mb4_unicode_ci,
  `label` mediumtext COLLATE utf8mb4_unicode_ci,
  `target` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



# Dump of table users
# ------------------------------------------------------------

CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `features` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '[]',
  `level` tinyint(1) NOT NULL DEFAULT '0',
  `canFeedback` tinyint(1) NOT NULL DEFAULT '0',
  `canAnnouncements` tinyint(1) NOT NULL DEFAULT '0',
  `showMigrateMessage` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
