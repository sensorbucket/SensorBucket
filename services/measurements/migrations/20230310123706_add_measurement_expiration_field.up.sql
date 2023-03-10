ALTER TABLE "measurements" ADD COLUMN measurement_expiration DATE NOT NULL DEFAULT(NOW() + '7 day'::interval);
