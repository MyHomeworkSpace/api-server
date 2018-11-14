# ************************************************************
# Sequel Pro SQL dump
# Version 4541
#
# http://www.sequelpro.com/
# https://github.com/sequelpro/sequelpro
#
# Host: 127.0.0.1 (MySQL 5.7.18)
# Database: myhomeworkspace
# Generation Time: 2018-11-14 02:59:27 +0000
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
  `text` text NOT NULL,
  `grade` int(2) NOT NULL,
  `type` int(2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table application_authorizations
# ------------------------------------------------------------

CREATE TABLE `application_authorizations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `applicationId` int(11) DEFAULT NULL,
  `token` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table applications
# ------------------------------------------------------------

CREATE TABLE `applications` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `name` text,
  `authorName` text,
  `clientId` text,
  `callbackUrl` text,
  `cors` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table calendar_classes
# ------------------------------------------------------------

CREATE TABLE `calendar_classes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `termId` int(11) DEFAULT NULL,
  `ownerId` int(11) DEFAULT NULL,
  `sectionId` int(11) DEFAULT NULL,
  `name` text,
  `ownerName` text,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table calendar_events
# ------------------------------------------------------------

CREATE TABLE `calendar_events` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `desc` text,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table calendar_hwevents
# ------------------------------------------------------------

CREATE TABLE `calendar_hwevents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `homeworkId` int(11) DEFAULT NULL,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table calendar_periods
# ------------------------------------------------------------

CREATE TABLE `calendar_periods` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `classId` int(11) DEFAULT NULL,
  `dayNumber` int(11) DEFAULT NULL,
  `block` text,
  `buildingName` text,
  `roomNumber` text,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table calendar_status
# ------------------------------------------------------------

CREATE TABLE `calendar_status` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table calendar_terms
# ------------------------------------------------------------

CREATE TABLE `calendar_terms` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `termId` int(11) DEFAULT NULL,
  `name` text,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table classes
# ------------------------------------------------------------

CREATE TABLE `classes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text,
  `teacher` text,
  `color` varchar(12) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table faculty
# ------------------------------------------------------------

CREATE TABLE `faculty` (
  `bbId` int(11) NOT NULL,
  `firstName` text NOT NULL,
  `lastName` text NOT NULL,
  `largeFileName` text NOT NULL,
  `department` text NOT NULL,
  `grades` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table faculty_periods
# ------------------------------------------------------------

CREATE TABLE `faculty_periods` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text,
  `sectionId` int(11) DEFAULT NULL,
  `room` text,
  `block` text,
  `dayNumber` int(11) DEFAULT NULL,
  `grade` int(11) DEFAULT NULL,
  `term` int(11) DEFAULT NULL,
  `start` int(11) DEFAULT NULL,
  `end` int(11) DEFAULT NULL,
  `facultyId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table feedback
# ------------------------------------------------------------

CREATE TABLE `feedback` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `type` varchar(10) DEFAULT NULL,
  `text` text,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table fridays
# ------------------------------------------------------------

CREATE TABLE `fridays` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `index` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table homework
# ------------------------------------------------------------

CREATE TABLE `homework` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text,
  `due` date DEFAULT NULL,
  `desc` text,
  `complete` varchar(45) DEFAULT NULL,
  `classId` int(11) DEFAULT NULL,
  `userId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table notifications
# ------------------------------------------------------------

CREATE TABLE `notifications` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `content` text,
  `expiry` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table prefs
# ------------------------------------------------------------

CREATE TABLE `prefs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `key` text,
  `value` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table sessions
# ------------------------------------------------------------

CREATE TABLE `sessions` (
  `id` varchar(255) NOT NULL,
  `userId` int(11) DEFAULT NULL,
  `userName` text,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;



# Dump of table tab_permissions
# ------------------------------------------------------------

CREATE TABLE `tab_permissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `tabId` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table tabs
# ------------------------------------------------------------

CREATE TABLE `tabs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `slug` text,
  `icon` text,
  `label` text,
  `target` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;



# Dump of table users
# ------------------------------------------------------------

CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  `username` text NOT NULL,
  `email` text NOT NULL,
  `type` text NOT NULL,
  `features` varchar(255) NOT NULL DEFAULT '[]',
  `level` tinyint(1) NOT NULL DEFAULT '0',
  `canFeedback` tinyint(1) NOT NULL DEFAULT '0',
  `canAnnouncements` tinyint(1) NOT NULL DEFAULT '0',
  `showMigrateMessage` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
