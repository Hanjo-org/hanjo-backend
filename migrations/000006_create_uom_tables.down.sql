-- ==============================================================================
-- Migration: 000006_create_uom_tables.down.sql
-- Module: UoM (Unit of Measurement) & Global Conversions
-- Description: Drops tables created in 000006_create_uom_tables.up.sql in reverse order.
-- ==============================================================================

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS uom_conversions CASCADE;
DROP TABLE IF EXISTS uoms CASCADE;
