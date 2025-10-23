<?php
// PHP example
$config = [
    'host' => 'localhost',
    'port' => 27017
];
// end-init

function connect($config) {
    return new MongoDB\Client("mongodb://{$config['host']}:{$config['port']}");
}

