# haproxy-dos-monitor
  Haproxy supports ratelimit per IP for specific URLs.
  Example:
  Haproxy is configured to blacklist the ips which accessed the specific urls(/signin) more than 5 times in 60s.
  But still it will affect the haproxy performance, so I am blacklisting those IPs in iptables based on haproxy ratelimit reports.
  This script will monitor the haproxy stats socket and blacklists the DOS IPs as per haproxy ratelimit configuration and sends the alert to slack.

# Requirements
   It requires environment variable SLACK_URL, we need to provide slack webhook url.
