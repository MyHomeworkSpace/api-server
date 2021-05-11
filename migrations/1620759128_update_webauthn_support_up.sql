-- Description: update webauthn support
-- Up migration
UPDATE `2fa`
SET `webauthn` = '1';
ALTER TABLE `2fa` CHANGE `webauthn` `webauthn` VARCHAR(2) DEFAULT '';
CREATE TABLE webauthn (
	`id` INT NOT NULL AUTO_INCREMENT,
	`userId` INT,
	`publicKey` BLOB,
	`AAGUID` BLOB,
	`signCount` INT(32),
	`cloneWarning` TINYINT(1),
	PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;