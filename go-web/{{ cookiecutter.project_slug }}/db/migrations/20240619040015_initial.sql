-- Create "authors" table
CREATE TABLE "authors" (
 "id" integer NOT NULL,
 "name" text NOT NULL,
 "bio" text NULL,
 PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "users" (
 "id" serial NOT NULL,
 "name" text NOT NULL,
 PRIMARY KEY ("id")
);
