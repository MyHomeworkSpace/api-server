-- Description: update webauthn support
-- Down migration
ALTER TABLE `2fa` CHANGE `webauthn` `webauthn` VARCHAR(255);
DROP TABLE webauthn;