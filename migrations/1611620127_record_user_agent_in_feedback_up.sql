-- Description: Record user agent in feedback
-- Up migration
ALTER TABLE feedback ADD userAgent varchar(255) NOT NULL DEFAULT '';
ALTER TABLE feedback CHANGE `userAgent` `userAgent` varchar(255) NOT NULL;