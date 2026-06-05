import yaml
import json

spec = yaml.safe_load(open('docs/api/openapi.yaml'))
paths = spec['paths']
components = spec.get('components', {})
schemas = components.get('schemas', {})

# Map OpenAPI tag -> friendly folder name
TAG_MAP = {
    'Health': 'Health',
    'Auth': 'Auth',
    'Users': 'Users',
    'Organizations': 'Organizations',
    'Properties': 'Properties',
    'UnitTypes': 'Unit Types',
    'Units': 'Units',
    'Amenities': 'Amenities',
    'Services': 'Services',
    'Pricing': 'Pricing',
    'RatePlans': 'Rate Plans',
    'Reservations': 'Reservations',
    'Availability': 'Availability',
    'Admin': 'Admin / DLQ',
}

def example_for_schema(name, depth=0):
    if depth > 3:
        return {}
    sch = schemas.get(name, {})
    if not sch:
        return {}
    if 'example' in sch:
        return sch['example']
    if '$ref' in sch:
        ref = sch['$ref'].split('/')[-1]
        return example_for_schema(ref, depth+1)
    t = sch.get('type', 'object')
    if t == 'object':
        out = {}
        props = sch.get('properties', {})
        for k, v in props.items():
            if '$ref' in v:
                out[k] = example_for_schema(v['$ref'].split('/')[-1], depth+1)
            elif v.get('type') == 'string':
                fmt = v.get('format', '')
                if fmt == 'uuid':
                    out[k] = '00000000-0000-0000-0000-000000000000'
                elif fmt == 'date':
                    out[k] = '2026-06-05'
                elif fmt == 'date-time':
                    out[k] = '2026-06-05T00:00:00Z'
                elif 'email' in k.lower():
                    out[k] = 'user@example.com'
                elif 'password' in k.lower():
                    out[k] = 'Good.Pass1'
                else:
                    out[k] = f'sample-{k}'
            elif v.get('type') == 'integer':
                out[k] = 1
            elif v.get('type') == 'number':
                out[k] = 0.0
            elif v.get('type') == 'boolean':
                out[k] = True
            elif v.get('type') == 'array':
                items = v.get('items', {})
                if '$ref' in items:
                    out[k] = [example_for_schema(items['$ref'].split('/')[-1], depth+1)]
                else:
                    out[k] = []
            else:
                out[k] = None
        return out
    return {}

def body_for_request(req_body):
    if not req_body:
        return None
    content = req_body.get('content', {})
    json_content = content.get('application/json', {})
    schema = json_content.get('schema', {})
    if '$ref' in schema:
        ref = schema['$ref'].split('/')[-1]
        return example_for_schema(ref)
    return example_for_schema(schema.get('title', ''))

# Build items by tag
folders = {}
folder_order = []

for path, methods in paths.items():
    for method, op in methods.items():
        if method not in ('get','post','put','delete','patch'):
            continue

        # Determine folder from tag
        tags = op.get('tags', [])
        tag = tags[0] if tags else 'Other'
        folder_name = TAG_MAP.get(tag, tag)
        if folder_name not in folders:
            folders[folder_name] = []
            folder_order.append(folder_name)

        # URL with base + path
        url_path = path
        # Path variables for Postman
        path_vars = []
        for part in url_path.split('/'):
            if part.startswith('{') and part.endswith('}'):
                pname = part[1:-1]
                path_vars.append({
                    'key': pname,
                    'value': '00000000-0000-0000-0000-000000000000' if pname in ('id', 'code') else 'value',
                    'description': f'Path parameter: {pname}'
                })

        request = {
            'method': method.upper(),
            'header': [
                {'key': 'Content-Type', 'value': 'application/json'},
                {'key': 'Accept', 'value': 'application/json'}
            ],
            'url': {
                'raw': '{{base_url}}' + url_path,
                'host': ['{{base_url}}'],
                'path': [p for p in url_path.split('/') if p and not p.startswith('{')] or ['/'],
            },
            'description': (op.get('description') or op.get('summary', '')).strip()
        }
        if path_vars:
            request['url']['variable'] = path_vars

        # Auth: most routes need bearer; if `security: []` is set, leave empty
        if op.get('security') == []:
            request['auth'] = {'type': 'noauth'}

        # Body
        body = body_for_request(op.get('requestBody'))
        if body is not None and method in ('post', 'put', 'patch'):
            request['body'] = {
                'mode': 'raw',
                'raw': json.dumps(body, indent=2),
                'options': {'raw': {'language': 'json'}}
            }

        # Add a test script for the login endpoint
        if path == '/api/v1/auth/login' and method == 'post':
            request['event'] = [{
                'listen': 'test',
                'script': {
                    'type': 'text/javascript',
                    'exec': [
                        "if (pm.response.code === 200) {",
                        "    const json = pm.response.json();",
                        "    if (json.token) {",
                        "        pm.collectionVariables.set('auth_token', json.token);",
                        "        console.log('Auth token saved to {{auth_token}}');",
                        "    }",
                        "}"
                    ]
                }
            }]

        folders[folder_name].append({
            'name': f'{method.upper()} {path}',
            'request': request
        })

folder_list = []
for name in folder_order:
    folder_list.append({
        'name': name,
        'description': f'{name} endpoints',
        'item': folders[name]
    })

collection = {
    'info': {
        'name': 'PMS Backend',
        'description': '''# PMS Backend API

Property Management System REST API.

**Authentication**: All `/api/v1/*` routes except `/api/v1/auth/*` require `Authorization: Bearer <jwt>`. Admin routes (`/admin/*`) require the SUPER_ADMIN role.

**Rate limiting**: `/api/v1/auth/*` is limited to 5 req/s with burst 10 per IP.

**Source**: This collection is generated from `docs/api/openapi.yaml`. To regenerate, run `python3 scripts/build_postman.py` (or the script you used).''',
        'schema': 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json'
    },
    'auth': {
        'type': 'bearer',
        'bearer': [{'key': 'token', 'value': '{{auth_token}}', 'type': 'string'}]
    },
    'variable': [
        {'key': 'base_url', 'value': 'http://localhost:8080', 'type': 'string',
         'description': 'Base URL of the API. Use http://localhost:8080 for local dev, or your staging URL.'},
        {'key': 'auth_token', 'value': '', 'type': 'string',
         'description': 'JWT obtained from POST /api/v1/auth/login. Set automatically by the login request test script.'}
    ],
    'item': folder_list
}

out_path = 'docs/api/pms-backend.postman_collection.json'
with open(out_path, 'w') as f:
    json.dump(collection, f, indent=2)

total = sum(len(f['item']) for f in folder_list)
print(f'Wrote {out_path}: {total} requests across {len(folder_list)} folders')
