mysql> show tables;
+-------------------------------+
| Tables_in_qmplus              |
+-------------------------------+
| authority_menu                |
| casbin_rule                   |
| exa_customers                 |
| exa_file_chunks               |
| exa_file_upload_and_downloads |
| exa_files                     |
| exa_simple_uploaders          |
| jwt_blacklists                |
| sys_apis                      |
| sys_authorities               |
| sys_authority_menus           |
| sys_base_menu_parameters      |
| sys_base_menus                |
| sys_data_authority_id         |
| sys_dictionaries              |
| sys_dictionary_details        |
| sys_operation_records         |
| sys_users                     |
| workflow_edges                |
| workflow_end_points           |
| workflow_nodes                |
| workflow_processes            |
| workflow_start_points         |
+-------------------------------+


drop view authority_menu;
drop table casbin_rule;
drop table exa_wf_leaves;
drop table exa_customers;
drop table  exa_file_chunks;
drop table exa_file_upload_and_downloads;
drop table exa_files;
drop table exa_simple_uploaders;
drop table jwt_blacklists;
drop table sys_apis;
drop table sys_authorities;
drop table sys_authority_menus;
drop table sys_base_menu_parameters;
drop table sys_base_menus;
drop table sys_data_authority_id;
drop table sys_dictionaries;
drop table sys_dictionary_details;
drop table sys_operation_records;
drop table sys_users;
drop table workflow_edges;
drop table workflow_end_points;
drop table workflow_nodes;
drop table workflow_processes;
drop table workflow_start_points;

