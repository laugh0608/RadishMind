DROP TRIGGER IF EXISTS gateway_model_pricing_revisions_no_update ON gateway_model_pricing_revisions;
DROP FUNCTION IF EXISTS gateway_model_pricing_revisions_append_only();
DROP TABLE IF EXISTS gateway_model_pricing_current;
DROP TABLE IF EXISTS gateway_model_pricing_revisions;
