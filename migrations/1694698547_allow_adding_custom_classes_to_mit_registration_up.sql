-- Description: Allow adding custom classes to MIT registration
-- Up migration

ALTER TABLE `mit_classes`
ADD `custom` tinyint(1) NOT NULL AFTER `sections`;