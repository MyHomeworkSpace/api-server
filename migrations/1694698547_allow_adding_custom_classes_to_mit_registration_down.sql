-- Description: Allow adding custom classes to MIT registration
-- Down migration

ALTER TABLE `mit_classes`
DROP `custom`;