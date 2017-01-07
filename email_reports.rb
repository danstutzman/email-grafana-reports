#!/usr/bin/ruby
require 'json'
require 'net/http'
require 'net/smtp'
require 'pp'
require 'rexml/document'

end_time   = Time.now.to_i
start_time = end_time - 60 * 60 * 24
Y_AXIS_SCALE = 1000.0
LEFT_MARGIN = 100

uri = URI('http://localhost:9090/api/v1/query_range')
uri.query = URI.encode_www_form({
  query: 'cloudfront_visits{site_name="vocabincontext.com",status="200"}',
  start: start_time,
  end:   end_time,
  step:  '20m',
})
response = Net::HTTP.get_response uri
if !response.is_a? Net::HTTPSuccess
  STDERR.puts "#{response.code} #{response.body}"
  exit 1
end
json = JSON.parse(response.body)
hour_percents = json['data']['result'][0]['values'].map { |pair|
  [(pair[0] - start_time) / 3600, pair[1].to_i]
}

doc = REXML::Document.new
svg = doc.add_element 'svg', {
  'viewBox' => '0 0 1000 200',
  'class' => 'chart',
}
svg.add_element 'polyline', {
  'fill' => 'none',
  'stroke' => '#0074d9',
  'stroke-width' => 3,
  'points' => hour_percents.map { |hour, percent|
    "#{LEFT_MARGIN + hour * 5},#{200 - percent / Y_AXIS_SCALE}"
  }.join(' '),
}
hour_percents.each do |hour, percent|
  if hour % 5 == 0
    text = svg.add_element 'text', {
      'x' => LEFT_MARGIN + hour * 5,
      'y' => 200,
    }
    text.text = hour
  end
end
(0...10).each do |y|
  value = y * 20 * Y_AXIS_SCALE
  text = svg.add_element 'text', {
    'x' => LEFT_MARGIN,
    'y' => 200 - (value / Y_AXIS_SCALE).to_i,
    'text-anchor' => 'end',
  }
  text.text = value
end

puts doc

Net::SMTP.start 'localhost' do |smtp|
  smtp.send_message "From: Reports <reports@monitoring.danstutzman.com>
To: Dan Stutzman <dtstutz@gmail.com>
MIME-Version: 1.0
Content-type: text/html
Subject: test subject" + "\n\n<html>#{doc}</html>",
    'reports@monitoring.danstutzman.com',
    'dtstutz@gmail.com'
end
