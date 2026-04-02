DO $$
BEGIN
  BEGIN
    ALTER TYPE laundry_backend."WorkflowStatus" ADD VALUE IF NOT EXISTS 'delivered' AFTER 'finished';
  EXCEPTION
    WHEN duplicate_object THEN
      NULL;
  END;
END $$;

