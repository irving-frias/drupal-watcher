#!/usr/bin/env php
<?php
/**
 * Drupal Watcher — binary downloader
 *
 * Called via:
 *   - Composer post-install-cmd / post-update-cmd
 *   - bin/drupal-watcher (PHP launcher, via require)
 *
 * Uses return instead of exit so it can be safely require'd.
 */

$installDir = __DIR__;
$packageJson = $installDir . '/../composer.json';
$binaryPath = $installDir . '/drupal-watcher-go';

$vendorBin = getenv('COMPOSER_RUNTIME_BIN_DIR');
if ($vendorBin === false || !is_dir($vendorBin)) {
	$vendorBin = null;
	$dir = $installDir;
	for ($i = 0; $i < 10; $i++) {
		$candidate = realpath($dir . '/vendor/bin');
		if ($candidate !== false && is_dir($candidate)) {
			$vendorBin = $candidate;
			break;
		}
		$parent = dirname($dir);
		if ($parent === $dir) break;
		$dir = $parent;
	}
}

if (!file_exists($packageJson)) {
	if ($vendorBin) {
		$proxyPath = $vendorBin . '/drupal-watcher';
		if (is_file($proxyPath) || is_link($proxyPath)) {
			@unlink($proxyPath);
		}
	}
	if (file_exists($binaryPath)) {
		@unlink($binaryPath);
	}
	return;
}

$composerMeta = json_decode(file_get_contents($packageJson), true);
$expectedVersion = $composerMeta['extra']['drupal-watcher-version'] ?? '1.0.0';

$needsDownload = true;
if (file_exists($binaryPath) && is_executable($binaryPath)) {
	$versionFile = $installDir . '/.binary-version';
	if (file_exists($versionFile) && trim(file_get_contents($versionFile)) === $expectedVersion) {
		$needsDownload = false;
	}
}

if (!$needsDownload) {
	return;
}

$repo = 'irving-frias/drupal-watcher';

$version = getenv('DRUPAL_WATCHER_VERSION');
if (!$version) {
	$apiUrl = "https://api.github.com/repos/{$repo}/releases/latest";
	$apiCtx = stream_context_create(['http' => ['header' => "User-Agent: Composer\r\n", 'timeout' => 5]]);
	$apiData = @file_get_contents($apiUrl, false, $apiCtx);
	if ($apiData) {
		$release = json_decode($apiData, true);
		if (isset($release['tag_name'])) {
			$version = $release['tag_name'];
		}
	}
}
if (!$version) {
	$version = $expectedVersion;
}

$osMap = [
	'Linux'    => 'linux',
	'Darwin'   => 'darwin',
	'WINNT'    => 'windows',
	'CYGWIN'   => 'windows',
	'FreeBSD'  => 'freebsd',
];
$goos = $osMap[PHP_OS] ?? strtolower(PHP_OS);

$archMap = [
	'x86_64'  => 'amd64',
	'amd64'   => 'amd64',
	'aarch64' => 'arm64',
	'arm64'   => 'arm64',
	'x86'     => '386',
	'i386'    => '386',
	'i686'    => '386',
];
$goarch = $archMap[php_uname('m')] ?? php_uname('m');

$binaryName = "drupal-watcher-{$goos}-{$goarch}";
$isWindows = $goos === 'windows';

if ($isWindows) {
	$archiveName = $binaryName . '.exe.zip';
} else {
	$archiveName = $binaryName . '.gz';
}

$url = "https://github.com/{$repo}/releases/download/{$version}/{$archiveName}";
echo "Downloading {$archiveName}...\n";

$context = stream_context_create([
	'http' => [
		'method' => 'GET',
		'timeout' => 30,
		'follow_location' => 1,
		'header' => "Accept: application/octet-stream\r\n",
	],
	'ssl' => [
		'verify_peer' => true,
		'verify_peer_name' => true,
	],
]);

$data = @file_get_contents($url, false, $context);
if ($data === false) {
	echo "Warning: download failed from {$url}\n";
	echo "To resolve: install Go locally or download the binary manually from:\n";
	echo "  https://github.com/{$repo}/releases\n";
	return;
}

if ($isWindows) {
	if (!class_exists('ZipArchive')) {
		echo "Error: Zip extension required. Unzip manually from: {$url}\n";
		return;
	}
	$zipPath = $installDir . '/drupal-watcher-tmp.zip';
	file_put_contents($zipPath, $data);
	$zip = new ZipArchive;
	if ($zip->open($zipPath) === true) {
		$zip->extractTo($installDir, ['drupal-watcher-windows-amd64.exe']);
		rename($installDir . '/drupal-watcher-windows-amd64.exe', $binaryPath);
		$zip->close();
	}
	unlink($zipPath);
} else {
	$decompressed = gzdecode($data);
	if ($decompressed === false) {
		echo "Error: failed to decompress {$archiveName}\n";
		return;
	}
	file_put_contents($binaryPath, $decompressed);
}

chmod($binaryPath, 0755);

file_put_contents($installDir . '/.binary-version', $version);

echo "Installed drupal-watcher {$version} ({$goos}/{$goarch})\n";
