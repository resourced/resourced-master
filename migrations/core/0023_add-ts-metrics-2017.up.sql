create table ts_metrics_m1_2017
    (check (created at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-01-31' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m2_2017
    (check (created at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-02-28' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m3_2017
    (check (created at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-03-31' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m4_2017
    (check (created at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-04-30' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m5_2017
    (check (created at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-05-31' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m6_2017
    (check (created at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-06-30' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m7_2017
    (check (created at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-07-31' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m8_2017
    (check (created at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-08-31' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m9_2017
    (check (created at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-09-30' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m10_2017
    (check (created at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-10-31' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m11_2017
    (check (created at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-11-30' at time zone 'utc'))
    inherits (ts_metrics);

create table ts_metrics_m12_2017
    (check (created at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-12-31' at time zone 'utc'))
    inherits (ts_metrics);

create index idx_ts_metrics_m1_2017_simple_select on ts_metrics_m1_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m2_2017_simple_select on ts_metrics_m2_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m3_2017_simple_select on ts_metrics_m3_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m4_2017_simple_select on ts_metrics_m4_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m5_2017_simple_select on ts_metrics_m5_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m6_2017_simple_select on ts_metrics_m6_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m7_2017_simple_select on ts_metrics_m7_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m8_2017_simple_select on ts_metrics_m8_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m9_2017_simple_select on ts_metrics_m9_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m10_2017_simple_select on ts_metrics_m10_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m11_2017_simple_select on ts_metrics_m11_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m12_2017_simple_select on ts_metrics_m12_2017 using brin (cluster_id, metric_id, created);

create index idx_ts_metrics_m1_2017_aggregate_select on ts_metrics_m1_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m2_2017_aggregate_select on ts_metrics_m2_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m3_2017_aggregate_select on ts_metrics_m3_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m4_2017_aggregate_select on ts_metrics_m4_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m5_2017_aggregate_select on ts_metrics_m5_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m6_2017_aggregate_select on ts_metrics_m6_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m7_2017_aggregate_select on ts_metrics_m7_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m9_2017_aggregate_select on ts_metrics_m9_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m10_2017_aggregate_select on ts_metrics_m10_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m11_2017_aggregate_select on ts_metrics_m11_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_m12_2017_aggregate_select on ts_metrics_m12_2017 using brin (cluster_id, created, key);


create or replace function on_ts_metrics_insert_2017() returns trigger as $$
begin
    if ( new.created at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-01-31' at time zone 'utc') then
        insert into ts_metrics_m1_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-02-28' at time zone 'utc') then
        insert into ts_metrics_m2_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-03-31' at time zone 'utc') then
        insert into ts_metrics_m3_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-04-30' at time zone 'utc') then
        insert into ts_metrics_m4_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-05-31' at time zone 'utc') then
        insert into ts_metrics_m5_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-06-30' at time zone 'utc') then
        insert into ts_metrics_m6_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-07-31' at time zone 'utc') then
        insert into ts_metrics_m7_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-08-31' at time zone 'utc') then
        insert into ts_metrics_m8_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-09-30' at time zone 'utc') then
        insert into ts_metrics_m9_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-10-31' at time zone 'utc') then
        insert into ts_metrics_m10_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-11-30' at time zone 'utc') then
        insert into ts_metrics_m11_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-12-31' at time zone 'utc') then
        insert into ts_metrics_m12_2017 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_metrics_insert_2017
    before insert on ts_metrics
    for each row execute procedure on_ts_metrics_insert_2017();


-- 15 minutes aggregate table
create table ts_metrics_aggr_15m_m1_2017
    (check (created at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-01-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m2_2017
    (check (created at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-02-28' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m3_2017
    (check (created at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-03-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m4_2017
    (check (created at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-04-30' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m5_2017
    (check (created at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-05-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m6_2017
    (check (created at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-06-30' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m7_2017
    (check (created at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-07-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m8_2017
    (check (created at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-08-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m9_2017
    (check (created at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-09-30' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m10_2017
    (check (created at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-10-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m11_2017
    (check (created at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-11-30' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m12_2017
    (check (created at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and created at time zone 'utc' <= date '2017-12-31' at time zone 'utc'))
    inherits (ts_metrics_aggr_15m);

create index idx_ts_metrics_aggr_15m_m1_2017_simple_select on ts_metrics_aggr_15m_m1_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m2_2017_simple_select on ts_metrics_aggr_15m_m2_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m3_2017_simple_select on ts_metrics_aggr_15m_m3_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m4_2017_simple_select on ts_metrics_aggr_15m_m4_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m5_2017_simple_select on ts_metrics_aggr_15m_m5_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m6_2017_simple_select on ts_metrics_aggr_15m_m6_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m7_2017_simple_select on ts_metrics_aggr_15m_m7_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m8_2017_simple_select on ts_metrics_aggr_15m_m8_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m9_2017_simple_select on ts_metrics_aggr_15m_m9_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m10_2017_simple_select on ts_metrics_aggr_15m_m10_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m11_2017_simple_select on ts_metrics_aggr_15m_m11_2017 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m12_2017_simple_select on ts_metrics_aggr_15m_m12_2017 using brin (cluster_id, metric_id, created);

create index idx_ts_metrics_aggr_15m_m1_2017_aggregate_select on ts_metrics_aggr_15m_m1_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m2_2017_aggregate_select on ts_metrics_aggr_15m_m2_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m3_2017_aggregate_select on ts_metrics_aggr_15m_m3_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m4_2017_aggregate_select on ts_metrics_aggr_15m_m4_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m5_2017_aggregate_select on ts_metrics_aggr_15m_m5_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m6_2017_aggregate_select on ts_metrics_aggr_15m_m6_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m7_2017_aggregate_select on ts_metrics_aggr_15m_m7_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m8_2017_aggregate_select on ts_metrics_aggr_15m_m8_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m9_2017_aggregate_select on ts_metrics_aggr_15m_m9_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m10_2017_aggregate_select on ts_metrics_aggr_15m_m10_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m11_2017_aggregate_select on ts_metrics_aggr_15m_m11_2017 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m12_2017_aggregate_select on ts_metrics_aggr_15m_m12_2017 using brin (cluster_id, created, key);


create or replace function on_ts_metrics_aggr_15m_insert_2017() returns trigger as $$
begin
    if ( new.created at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-01-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m1_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-02-28' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m2_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-03-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m3_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-04-30' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m4_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-05-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m5_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-06-30' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m6_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-07-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m7_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-08-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m8_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-09-30' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m9_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-10-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m10_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-11-30' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m11_2017 values (new.*);
    elsif ( new.created at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and new.created at time zone 'utc' <= date '2017-12-31' at time zone 'utc') then
        insert into ts_metrics_aggr_15m_m12_2017 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_metrics_aggr_15m_insert_2017
    before insert on ts_metrics_aggr_15m
    for each row execute procedure on_ts_metrics_aggr_15m_insert_2017();
