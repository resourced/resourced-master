create table ts_events_m1_2016
    (check (created_from >= date '2016-01-01' and created_from <= date '2016-01-31'))
    inherits (ts_events);

create table ts_events_m2_2016
    (check (created_from >= date '2016-02-01' and created_from <= date '2016-02-29'))
    inherits (ts_events);

create table ts_events_m3_2016
    (check (created_from >= date '2016-03-01' and created_from <= date '2016-03-31'))
    inherits (ts_events);

create table ts_events_m4_2016
    (check (created_from >= date '2016-04-01' and created_from <= date '2016-04-30'))
    inherits (ts_events);

create table ts_events_m5_2016
    (check (created_from >= date '2016-05-01' and created_from <= date '2016-05-31'))
    inherits (ts_events);

create table ts_events_m6_2016
    (check (created_from >= date '2016-06-01' and created_from <= date '2016-06-30'))
    inherits (ts_events);

create table ts_events_m7_2016
    (check (created_from >= date '2016-07-01' and created_from <= date '2016-07-31'))
    inherits (ts_events);

create table ts_events_m8_2016
    (check (created_from >= date '2016-08-01' and created_from <= date '2016-08-31'))
    inherits (ts_events);

create table ts_events_m9_2016
    (check (created_from >= date '2016-09-01' and created_from <= date '2016-09-30'))
    inherits (ts_events);

create table ts_events_m10_2016
    (check (created_from >= date '2016-10-01' and created_from <= date '2016-10-31'))
    inherits (ts_events);

create table ts_events_m11_2016
    (check (created_from >= date '2016-11-01' and created_from <= date '2016-11-30'))
    inherits (ts_events);

create table ts_events_m12_2016
    (check (created_from >= date '2016-12-01' and created_from <= date '2016-12-31'))
    inherits (ts_events);

create index idx_ts_events_m1_2016_simple_select on ts_events_m1_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m2_2016_simple_select on ts_events_m2_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m3_2016_simple_select on ts_events_m3_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m4_2016_simple_select on ts_events_m4_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m5_2016_simple_select on ts_events_m5_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m6_2016_simple_select on ts_events_m6_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m7_2016_simple_select on ts_events_m7_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m8_2016_simple_select on ts_events_m8_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m9_2016_simple_select on ts_events_m9_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m10_2016_simple_select on ts_events_m10_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m11_2016_simple_select on ts_events_m11_2016 using brin (cluster_id, created_from, created_to);
create index idx_ts_events_m12_2016_simple_select on ts_events_m12_2016 using brin (cluster_id, created_from, created_to);

create unique index idx_ts_events_m1_2016_id on ts_events_m1_2016 (id);
create unique index idx_ts_events_m2_2016_id on ts_events_m2_2016 (id);
create unique index idx_ts_events_m3_2016_id on ts_events_m3_2016 (id);
create unique index idx_ts_events_m4_2016_id on ts_events_m4_2016 (id);
create unique index idx_ts_events_m5_2016_id on ts_events_m5_2016 (id);
create unique index idx_ts_events_m6_2016_id on ts_events_m6_2016 (id);
create unique index idx_ts_events_m7_2016_id on ts_events_m7_2016 (id);
create unique index idx_ts_events_m8_2016_id on ts_events_m8_2016 (id);
create unique index idx_ts_events_m9_2016_id on ts_events_m9_2016 (id);
create unique index idx_ts_events_m10_2016_id on ts_events_m10_2016 (id);
create unique index idx_ts_events_m11_2016_id on ts_events_m11_2016 (id);
create unique index idx_ts_events_m12_2016_id on ts_events_m12_2016 (id);

create or replace function on_ts_events_insert_2016() returns trigger as $$
begin
    if ( new.created_from >= date '2016-01-01' and new.created_from <= date '2016-01-31') then
        insert into ts_events_m1_2016 values (new.*);
    elsif ( new.created_from >= date '2016-02-01' and new.created_from <= date '2016-02-29') then
        insert into ts_events_m2_2016 values (new.*);
    elsif ( new.created_from >= date '2016-03-01' and new.created_from <= date '2016-03-31') then
        insert into ts_events_m3_2016 values (new.*);
    elsif ( new.created_from >= date '2016-04-01' and new.created_from <= date '2016-04-30') then
        insert into ts_events_m4_2016 values (new.*);
    elsif ( new.created_from >= date '2016-05-01' and new.created_from <= date '2016-05-31') then
        insert into ts_events_m5_2016 values (new.*);
    elsif ( new.created_from >= date '2016-06-01' and new.created_from <= date '2016-06-30') then
        insert into ts_events_m6_2016 values (new.*);
    elsif ( new.created_from >= date '2016-07-01' and new.created_from <= date '2016-07-31') then
        insert into ts_events_m7_2016 values (new.*);
    elsif ( new.created_from >= date '2016-08-01' and new.created_from <= date '2016-08-31') then
        insert into ts_events_m8_2016 values (new.*);
    elsif ( new.created_from >= date '2016-09-01' and new.created_from <= date '2016-09-30') then
        insert into ts_events_m9_2016 values (new.*);
    elsif ( new.created_from >= date '2016-10-01' and new.created_from <= date '2016-10-31') then
        insert into ts_events_m10_2016 values (new.*);
    elsif ( new.created_from >= date '2016-11-01' and new.created_from <= date '2016-11-30') then
        insert into ts_events_m11_2016 values (new.*);
    elsif ( new.created_from >= date '2016-12-01' and new.created_from <= date '2016-12-31') then
        insert into ts_events_m12_2016 values (new.*);
    else
        raise exception 'created_from date out of range';
    end if;

    return null;
end;
$$ language plpgsql;

create trigger ts_events_insert_2016
    before insert on ts_events
    for each row execute procedure on_ts_events_insert_2016();
