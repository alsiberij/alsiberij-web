<?php
ini_set('default_charset', '');

header_remove('X-Powered-By');

$uri = explode('?', $_SERVER['REQUEST_URI'])[0];

if (preg_match('~\.(css)$~', $uri)) {
    $path = './view/' . $uri;
    if (!file_exists($path)) {
        http_response_code(404);
    } else {
        header('Content-type: text/css');
        require_once $path;
    }
    return;
}
if (preg_match('~\.(png)$~', $uri)) {
    $path = './view/' . $uri;
    if (!file_exists($path)) {
        http_response_code(404);
    } else {
        header('Content-type: image/png');
        require_once $path;
    }
    return;
}

if ($uri == '/') {
    require_once('./view/html/index.html');
} else {
    require_once('./view/html/error.html');
}