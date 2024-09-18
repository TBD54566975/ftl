-- migrate:up
DROP INDEX deployments_unique_idx;
CREATE INDEX deployments_active_idx ON public.deployments USING btree (module_id) WHERE (min_replicas > 0);
-- migrate:down

