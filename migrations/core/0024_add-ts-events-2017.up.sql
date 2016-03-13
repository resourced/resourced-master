create table ts_events_m1_2017
    (check (created_from at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-01-31' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m2_2017
    (check (created_from at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-02-28' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m3_2017
    (check (created_from at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-03-31' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m4_2017
    (check (created_from at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-04-30' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m5_2017
    (check (created_from at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-05-31' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m6_2017
    (check (created_from at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-06-30' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m7_2017
    (check (created_from at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-07-31' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m8_2017
    (check (created_from at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-08-31' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m9_2017
    (check (created_from at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-09-30' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m10_2017
    (check (created_from at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-10-31' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m11_2017
    (check (created_from at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-11-30' at time zone 'utc'))
    inherits (ts_events);

create table ts_events_m12_2017
    (check (created_from at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and created_from at time zone 'utc' <= date '2017-12-31' at time zone 'utc'))
    inherits (ts_events);

create index idx_ts_events_m1_2017_simple_select on ts_events_m1_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m2_2017_simple_select on ts_events_m2_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m3_2017_simple_select on ts_events_m3_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m4_2017_simple_select on ts_events_m4_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m5_2017_simple_select on ts_events_m5_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m6_2017_simple_select on ts_events_m6_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m7_2017_simple_select on ts_events_m7_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m8_2017_simple_select on ts_events_m8_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m9_2017_simple_select on ts_events_m9_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m10_2017_simple_select on ts_events_m10_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m11_2017_simple_select on ts_events_m11_2017 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m12_2017_simple_select on ts_events_m12_2017 using brin (cluster_id, created_from, created_to);

create index idx_ts_events_m1_2017_id on ts_events_m1_2017 (id);
create index idx_ts_events_m2_2017_id on ts_events_m2_2017 (id);
create index idx_ts_events_m3_2017_id on ts_events_m3_2017 (id);
create index idx_ts_events_m4_2017_id on ts_events_m4_2017 (id);
create index idx_ts_events_m5_2017_id on ts_events_m5_2017 (id);
create index idx_ts_events_m6_2017_id on ts_events_m6_2017 (id);
create index idx_ts_events_m7_2017_id on ts_events_m7_2017 (id);
create index idx_ts_events_m8_2017_id on ts_events_m8_2017 (id);
create index idx_ts_events_m9_2017_id on ts_events_m9_2017 (id);
create index idx_ts_events_m10_2017_id on ts_events_m10_2017 (id);
create index idx_ts_events_m11_2017_id on ts_events_m11_2017 (id);
create index idx_ts_events_m12_2017_id on ts_events_m12_2017 (id);

create or replace function on_ts_events_insert_2017() returns trigger as $$
begin
    if ( new.created_from at time zone 'utc' >= date '2017-01-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-01-31' at time zone 'utc') then
        insert into ts_events_m1_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-02-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-02-28' at time zone 'utc') then
        insert into ts_events_m2_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-03-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-03-31' at time zone 'utc') then
        insert into ts_events_m3_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-04-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-04-30' at time zone 'utc') then
        insert into ts_events_m4_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-05-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-05-31' at time zone 'utc') then
        insert into ts_events_m5_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-06-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-06-30' at time zone 'utc') then
        insert into ts_events_m6_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-07-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-07-31' at time zone 'utc') then
        insert into ts_events_m7_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-08-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-08-31' at time zone 'utc') then
        insert into ts_events_m8_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-09-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-09-30' at time zone 'utc') then
        insert into ts_events_m9_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-10-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-10-31' at time zone 'utc') then
        insert into ts_events_m10_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-11-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-11-30' at time zone 'utc') then
        insert into ts_events_m11_2017 values (new.*);
    elsif ( new.created_from at time zone 'utc' >= date '2017-12-01' at time zone 'utc' and new.created_from at time zone 'utc' <= date '2017-12-31' at time zone 'utc') then
        insert into ts_events_m12_2017 values (new.*);
    else
        raise exception 'created_from date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_events_insert_2017
    before insert on ts_events
    for each row execute procedure on_ts_events_insert_2017();
