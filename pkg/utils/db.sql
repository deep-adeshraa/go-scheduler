CREATE TABLE
  public.personal_test_job_schedules (
    id serial NOT NULL,
    function character varying(255) NULL,
    picked_at timestamp without time zone NULL,
    scheduled_at timestamp with time zone NULL,
    started_at timestamp with time zone NULL,
    completed_at timestamp with time zone NULL,
    failed_at timestamp with time zone NULL
  );

ALTER TABLE
  public.personal_test_job_schedules
ADD
  CONSTRAINT personal_test_job_schedules_pkey PRIMARY KEY (id)