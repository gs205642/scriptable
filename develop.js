let file

function getServerIp() {
	if (Keychain.contains('joe_server_ip')) {
		return Keychain.get('joe_server_ip')
	}
	return '192.168.1.3'
}

async function setServerIp() {
	const alert = new Alert()
	alert.title = '服务器 IP'
	alert.message = '请输入远程开发服务器（电脑）IP地址'
	alert.addTextField('server ip', getServerIp())
	alert.addAction('连接')
	alert.addCancelAction('取消')

	await alert.present()
	Keychain.set('joe_server_ip', alert.textFieldValue(0))
}

function serverUrl(path) {
	return `http://${getServerIp()}:8080/${path}`
}

async function upload() {
	const r = new Request(serverUrl('upload'))
	r.method = 'POST'
	r.addFileToMultipart(file, 'file')
	return r.loadJSON().then(data => {
		if (!data.success) {
			throw data.message
		}
		return data.message
	})
}

function serverLog(message, type='log') {
	const r = new Request(serverUrl('console'))
	r.method = 'POST'
	r.headers = {
		'Content-Type': 'application/json'
	}
	r.body = JSON.stringify({
		type,
		message,
	})
	r.load()
}

function log(message) {
	serverLog(message)
	console.log(message)
}

function logWarning(message) {
	serverLog(message, 'warn')
	console.warn(message)
}

function logError(message) {
	serverLog(message, 'error')
	console.error(message)
}

async function sync(fileName) {
	const r = new Request(serverUrl(`sync/${fileName}`))
	r.timeoutInterval = 1800
	await r.loadJSON().then(async({success, message: data}) => {
		if (!success) {
			throw data
		}
		try {
			await eval(`(async() => {${data}})()`)
		} catch (error) {
			logError(`JS错误：${error}`)
		}
		FileManager.local().writeString(file, data)
	})
}

async function develop() {
	file = await DocumentPicker.openFile()
	try {
		const fileName = await upload()
		while (true) {
			await sync(fileName)
		}
	} catch (error) {
		logError(error)
	}
}

async function menu() {
	const menu = new Alert()
	menu.addAction('远程开发')
	menu.addAction('修改服务器地址')
	menu.addCancelAction('取消')

	return menu.presentSheet()
}

switch (await menu()) {
	case 0:
		await develop()
		break
	case 1:
		await setServerIp()
		break
	default:
		break
}
