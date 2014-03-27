-- First create a user:
--
-- useradd encore-analysis
-- createuser -D -e -l -R -S encore-analysis

GRANT SELECT ON ALL TABLES IN SCHEMA public TO "encore-analysis";
GRANT USAGE ON SCHEMA PUBLIC TO "encore-analysis";
CREATE SCHEMA analysis;
GRANT ALL ON ALL TABLES IN SCHEMA analysis TO "encore-analysis";
GRANT ALL ON SCHEMA analysis TO "encore-analysis";
