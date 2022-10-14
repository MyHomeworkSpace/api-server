-- Description: Add description and location to all events
-- Up migration

ALTER TABLE `calendar_events`
ADD `location` text NOT NULL AFTER `end`;

ALTER TABLE `calendar_hwevents`
ADD `location` text NOT NULL AFTER `end`,
ADD `desc` text NOT NULL AFTER `location`;