#!/usr/bin/python
# -*- coding:utf-8 -*-

import sys,os
import urllib, urllib2
import base64
import hmac
import hashlib
from hashlib import sha1
import time
import uuid
import json
from optparse import OptionParser
import ConfigParser
from string import Template

access_key_id = '';
access_key_secret = '';
endpoint = 'https://sts.aliyuncs.com'
DEFAULT_PROFILE_SECTION = 'Credentials'
APIVERSION='2015-04-01'

def percent_encode(str):
    res = urllib.quote(str.decode(sys.stdin.encoding).encode('utf8'), '')
    res = res.replace('+', '%20')
    res = res.replace('*', '%2A')
    res = res.replace('%7E', '~')
    return res

def compute_signature(parameters, access_key_secret):
    sortedParameters = sorted(parameters.items(), key=lambda parameters: parameters[0])

    canonicalizedQueryString = ''
    for (k,v) in sortedParameters:
        canonicalizedQueryString += '&' + percent_encode(k) + '=' + percent_encode(v)

    stringToSign = 'GET&%2F&' + percent_encode(canonicalizedQueryString[1:])

    h = hmac.new(access_key_secret + "&", stringToSign, sha1)
    signature = base64.encodestring(h.digest()).strip()
    return signature

def compose_url(user_params):
    timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())

    parameters = { 
            'Format'        : 'JSON', 
            'Version'       :  APIVERSION, 
            'AccessKeyId'   : access_key_id, 
            'SignatureVersion'  : '1.0', 
            'SignatureMethod'   : 'HMAC-SHA1', 
            'SignatureNonce'    : str(uuid.uuid1()), 
            'Timestamp'         : time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    }

    for key in user_params.keys():
        parameters[key] = user_params[key]

    signature = compute_signature(parameters, access_key_secret)
    parameters['Signature'] = signature
    url = endpoint + "/?" + urllib.urlencode(parameters)
    return url

def make_request(user_params, quiet=False):
    url = compose_url(user_params)
    print url
    request = urllib2.Request(url)

    try:
        conn = urllib2.urlopen(request)
        response = conn.read()
    except urllib2.HTTPError, e:
        print(e.read().strip())
        raise SystemExit(e)

    #make json output pretty, this code is copied from json.tool
    try:
        obj = json.loads(response)
        if quiet:
            return obj
    except ValueError, e:
        raise SystemExit(e)
    json.dump(obj, sys.stdout, sort_keys=True, indent=2)
    sys.stdout.write('\n')

def list_profiles():
    config = ConfigParser.ConfigParser()
    try:
        config.read(get_config_file_path())
        sections = config.sections()
        print 'Profiles:'
        for section in sections:
            print '    %s' % section
    except Exception, e:
        print("can't read config file")

def get_config_file_path():
    if(os.path.isfile('aliyuncredentials')):
        return 'aliyuncredentials'
    else:
        return os.path.expanduser('~') + '/.aliyuncredentials'

def configure_accesskeypair(args, options):
    if options.accesskeyid is None or options.accesskeysecret is None:
        print("config miss parameters, use --id=[accesskeyid] --secret=[accesskeysecret]")
        sys.exit(1)

    section = DEFAULT_PROFILE_SECTION
    if options.profile is not None and options.profile != '':
        section = options.profile

    config = ConfigParser.RawConfigParser()
    config.read(get_config_file_path())
    config.add_section(section)
    config.set(section, 'accesskeyid', options.accesskeyid)
    config.set(section, 'accesskeysecret', options.accesskeysecret)
    cfgfile = open(get_config_file_path(), 'w+')
    config.write(cfgfile)
    cfgfile.close()

def setup_credentials():
    try:
        global access_key_id
        global access_key_secret
        if(options.accesskeyid is not None or options.accesskeysecret is not None):
            access_key_id = options.accesskeyid
            access_key_secret = options.accesskeysecret
        else:
            config = ConfigParser.ConfigParser()
            config.read(get_config_file_path())
            section = DEFAULT_PROFILE_SECTION
            if options.profile is not None and options.profile != '':
                section = options.profile
            access_key_id = config.get(section, 'accesskeyid')
            access_key_secret = config.get(section, 'accesskeysecret')
    except ConfigParser.NoSectionError, e:
        print 'can not find the profile, ', e;
        sys.exit(1)

    except Exception, e:
        print("can't get access key pair, use config --id=[accesskeyid] --secret=[accesskeysecret] to setup")
        sys.exit(1)

if __name__ == '__main__':
    help_doc='''
    $file Action Param1=Value1 Param2=Value2
Example:
    $file GetUser UserName=test
    $file CreatePolicy PolicyName=test PolicyDocument=file:my_policy.json
'''
    parser = OptionParser(Template(help_doc).substitute(file=sys.argv[0]))
    parser.add_option("-p", "--profile", dest="profile", help="specify the profile")
    parser.add_option("-i", "--id", dest="accesskeyid", help="specify access key id")
    parser.add_option("-s", "--secret", dest="accesskeysecret", help="specify access key secret")
    parser.add_option("-v", "--version", dest="version", help="API Version")

    (options, args) = parser.parse_args()
    if len(args) < 1:
        parser.print_help()
        list_profiles()
        sys.exit(0)

    if args[0] != 'config':
        setup_credentials()
    else: #it's a configure id/secret command
        configure_accesskeypair(args, options)
        sys.exit(0)
    idx = 1
    if options.version is not None:
        APIVERSION=options.version
        idx += 2
    user_params = {}
    user_params['Action'] = sys.argv[idx]
    idx += 1
    for arg in sys.argv[idx:]:
        try:
            params = arg.split('=')
            if(len(params) != 2 or params[0][0:1] == '-'):
                continue

            user_params[params[0]] = params[1]
        except ValueError, e:
            print(e.read().strip())
            raise SystemExit(e)

    #make api call now, for a bug in describeInstanceStatus, we do a special operation
    if user_params.has_key('policyfile'):
        user_params['policydocument'] = file(user_params['policyfile']).read()
    for (key,value) in user_params.iteritems():
        if(value.lower()[0:5] == 'file:'):
            file_path = value[5:]
            f = open(file_path)
            content = f.read()
            f.close()
            user_params[key] = content

    make_request(user_params)

