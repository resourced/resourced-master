create table ts_metrics_m1_2016
    (check (created >= TIMESTAMPTZ '2016-01-01 00:00:00-00' and created < TIMESTAMPTZ '2016-02-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m2_2016
    (check (created >= TIMESTAMPTZ '2016-02-01 00:00:00-00' and created < TIMESTAMPTZ '2016-03-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m3_2016
    (check (created >= TIMESTAMPTZ '2016-03-01 00:00:00-00' and created < TIMESTAMPTZ '2016-04-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m4_2016
    (check (created >= TIMESTAMPTZ '2016-04-01 00:00:00-00' and created < TIMESTAMPTZ '2016-05-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m5_2016
    (check (created >= TIMESTAMPTZ '2016-05-01 00:00:00-00' and created < TIMESTAMPTZ '2016-06-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m6_2016
    (check (created >= TIMESTAMPTZ '2016-06-01 00:00:00-00' and created < TIMESTAMPTZ '2016-07-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m7_2016
    (check (created >= TIMESTAMPTZ '2016-07-01 00:00:00-00' and created < TIMESTAMPTZ '2016-08-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m8_2016
    (check (created >= TIMESTAMPTZ '2016-08-01 00:00:00-00' and created < TIMESTAMPTZ '2016-09-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m9_2016
    (check (created >= TIMESTAMPTZ '2016-09-01 00:00:00-00' and created < TIMESTAMPTZ '2016-10-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m10_2016
    (check (created >= TIMESTAMPTZ '2016-10-01 00:00:00-00' and created < TIMESTAMPTZ '2016-11-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m11_2016
    (check (created >= TIMESTAMPTZ '2016-11-01 00:00:00-00' and created < TIMESTAMPTZ '2016-12-01 00:00:00-00'))
    inherits (ts_metrics);

create table ts_metrics_m12_2016
    (check (created >= TIMESTAMPTZ '2016-12-01 00:00:00-00' and created < TIMESTAMPTZ '2017-01-01 00:00:00-00'))
    inherits (ts_metrics);

create index idx_ts_metrics_m1_2016_simple_select on ts_metrics_m1_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m2_2016_simple_select on ts_metrics_m2_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m3_2016_simple_select on ts_metrics_m3_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m4_2016_simple_select on ts_metrics_m4_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m5_2016_simple_select on ts_metrics_m5_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m6_2016_simple_select on ts_metrics_m6_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m7_2016_simple_select on ts_metrics_m7_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m8_2016_simple_select on ts_metrics_m8_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m9_2016_simple_select on ts_metrics_m9_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m10_2016_simple_select on ts_metrics_m10_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m11_2016_simple_select on ts_metrics_m11_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_m12_2016_simple_select on ts_metrics_m12_2016 using brin (cluster_id, metric_id, created);

create index idx_ts_metrics_m1_2016_aggregate_select on ts_metrics_m1_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m2_2016_aggregate_select on ts_metrics_m2_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m3_2016_aggregate_select on ts_metrics_m3_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m4_2016_aggregate_select on ts_metrics_m4_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m5_2016_aggregate_select on ts_metrics_m5_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m6_2016_aggregate_select on ts_metrics_m6_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m7_2016_aggregate_select on ts_metrics_m7_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m9_2016_aggregate_select on ts_metrics_m9_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m10_2016_aggregate_select on ts_metrics_m10_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m11_2016_aggregate_select on ts_metrics_m11_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_m12_2016_aggregate_select on ts_metrics_m12_2016 using brin (cluster_id, created, key);


create or replace function on_ts_metrics_insert_2016() returns trigger as $$
begin
    if ( new.created >= TIMESTAMPTZ '2016-01-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-02-01 00:00:00-00') then
        insert into ts_metrics_m1_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-02-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-03-01 00:00:00-00') then
        insert into ts_metrics_m2_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-03-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-04-01 00:00:00-00') then
        insert into ts_metrics_m3_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-04-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-05-01 00:00:00-00') then
        insert into ts_metrics_m4_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-05-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-06-01 00:00:00-00') then
        insert into ts_metrics_m5_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-06-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-07-01 00:00:00-00') then
        insert into ts_metrics_m6_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-07-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-08-01 00:00:00-00') then
        insert into ts_metrics_m7_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-08-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-09-01 00:00:00-00') then
        insert into ts_metrics_m8_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-09-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-10-01 00:00:00-00') then
        insert into ts_metrics_m9_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-10-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-11-01 00:00:00-00') then
        insert into ts_metrics_m10_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-11-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-12-01 00:00:00-00') then
        insert into ts_metrics_m11_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-12-01 00:00:00-00' and new.created < TIMESTAMPTZ '2017-01-01 00:00:00-00') then
        insert into ts_metrics_m12_2016 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_metrics_insert_2016
    before insert on ts_metrics
    for each row execute procedure on_ts_metrics_insert_2016();


-- 15 minutes aggregate table
create table ts_metrics_aggr_15m_m1_2016
    (check (created >= TIMESTAMPTZ '2016-01-01 00:00:00-00' and created < TIMESTAMPTZ '2016-02-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m2_2016
    (check (created >= TIMESTAMPTZ '2016-02-01 00:00:00-00' and created < TIMESTAMPTZ '2016-03-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m3_2016
    (check (created >= TIMESTAMPTZ '2016-03-01 00:00:00-00' and created < TIMESTAMPTZ '2016-04-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m4_2016
    (check (created >= TIMESTAMPTZ '2016-04-01 00:00:00-00' and created < TIMESTAMPTZ '2016-05-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m5_2016
    (check (created >= TIMESTAMPTZ '2016-05-01 00:00:00-00' and created < TIMESTAMPTZ '2016-06-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m6_2016
    (check (created >= TIMESTAMPTZ '2016-06-01 00:00:00-00' and created < TIMESTAMPTZ '2016-07-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m7_2016
    (check (created >= TIMESTAMPTZ '2016-07-01 00:00:00-00' and created < TIMESTAMPTZ '2016-08-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m8_2016
    (check (created >= TIMESTAMPTZ '2016-08-01 00:00:00-00' and created < TIMESTAMPTZ '2016-09-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m9_2016
    (check (created >= TIMESTAMPTZ '2016-09-01 00:00:00-00' and created < TIMESTAMPTZ '2016-10-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m10_2016
    (check (created >= TIMESTAMPTZ '2016-10-01 00:00:00-00' and created < TIMESTAMPTZ '2016-11-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m11_2016
    (check (created >= TIMESTAMPTZ '2016-11-01 00:00:00-00' and created < TIMESTAMPTZ '2016-12-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create table ts_metrics_aggr_15m_m12_2016
    (check (created >= TIMESTAMPTZ '2016-12-01 00:00:00-00' and created < TIMESTAMPTZ '2017-01-01 00:00:00-00'))
    inherits (ts_metrics_aggr_15m);

create index idx_ts_metrics_aggr_15m_m1_2016_simple_select on ts_metrics_aggr_15m_m1_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m2_2016_simple_select on ts_metrics_aggr_15m_m2_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m3_2016_simple_select on ts_metrics_aggr_15m_m3_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m4_2016_simple_select on ts_metrics_aggr_15m_m4_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m5_2016_simple_select on ts_metrics_aggr_15m_m5_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m6_2016_simple_select on ts_metrics_aggr_15m_m6_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m7_2016_simple_select on ts_metrics_aggr_15m_m7_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m8_2016_simple_select on ts_metrics_aggr_15m_m8_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m9_2016_simple_select on ts_metrics_aggr_15m_m9_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m10_2016_simple_select on ts_metrics_aggr_15m_m10_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m11_2016_simple_select on ts_metrics_aggr_15m_m11_2016 using brin (cluster_id, metric_id, created);
create index idx_ts_metrics_aggr_15m_m12_2016_simple_select on ts_metrics_aggr_15m_m12_2016 using brin (cluster_id, metric_id, created);

create index idx_ts_metrics_aggr_15m_m1_2016_aggregate_select on ts_metrics_aggr_15m_m1_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m2_2016_aggregate_select on ts_metrics_aggr_15m_m2_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m3_2016_aggregate_select on ts_metrics_aggr_15m_m3_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m4_2016_aggregate_select on ts_metrics_aggr_15m_m4_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m5_2016_aggregate_select on ts_metrics_aggr_15m_m5_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m6_2016_aggregate_select on ts_metrics_aggr_15m_m6_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m7_2016_aggregate_select on ts_metrics_aggr_15m_m7_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m8_2016_aggregate_select on ts_metrics_aggr_15m_m8_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m9_2016_aggregate_select on ts_metrics_aggr_15m_m9_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m10_2016_aggregate_select on ts_metrics_aggr_15m_m10_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m11_2016_aggregate_select on ts_metrics_aggr_15m_m11_2016 using brin (cluster_id, created, key);
create index idx_ts_metrics_aggr_15m_m12_2016_aggregate_select on ts_metrics_aggr_15m_m12_2016 using brin (cluster_id, created, key);


create or replace function on_ts_metrics_aggr_15m_insert_2016() returns trigger as $$
begin
    if ( new.created >= TIMESTAMPTZ '2016-01-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-02-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m1_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-02-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-03-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m2_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-03-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-04-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m3_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-04-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-05-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m4_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-05-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-06-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m5_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-06-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-07-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m6_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-07-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-08-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m7_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-08-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-09-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m8_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-09-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-10-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m9_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-10-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-11-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m10_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-11-01 00:00:00-00' and new.created < TIMESTAMPTZ '2016-12-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m11_2016 values (new.*);
    elsif ( new.created >= TIMESTAMPTZ '2016-12-01 00:00:00-00' and new.created < TIMESTAMPTZ '2017-01-01 00:00:00-00') then
        insert into ts_metrics_aggr_15m_m12_2016 values (new.*);
    else
        raise exception 'created date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_metrics_aggr_15m_insert_2016
    before insert on ts_metrics_aggr_15m
    for each row execute procedure on_ts_metrics_aggr_15m_insert_2016();
