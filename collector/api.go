package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ============ EVENT INGESTION ============

func eventHandler(w http.ResponseWriter, r *http.Request) {
	dev, ok := validateToken(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var e Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid event body", http.StatusBadRequest)
		return
	}

	e.DeviceID = dev.ID
	if e.Device == "" {
		e.Device = dev.Type
	}
	e.Tenant = dev.Tenant

	eventChan <- e
	w.Write([]byte("queued"))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	dev, ok := validateToken(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var e Event
		if err := json.Unmarshal(msg, &e); err != nil {
			continue
		}

		e.DeviceID = dev.ID
		if e.Device == "" {
			e.Device = dev.Type
		}
		e.Tenant = dev.Tenant

		eventChan <- e
	}
}

// ============ DASHBOARD UI ============

func dashboardHandler(w http.ResponseWriter, r *http.Request) {

	html := `
<!doctype html>
<html>
<head>
<title>Mini CDO Dashboard</title>
<style>

body {
  margin:0; padding:0;
  font-family: Arial;
  display:flex;
  height:100vh;
  background:#f0f2f5;
}

/* Sidebar */
#sidebar {
  width:280px;
  background:white;
  border-right:1px solid #ccc;
  padding:20px;
  overflow-y:auto;
}

.card {
  padding:10px;
  border-radius:8px;
  margin-bottom:10px;
  text-align:center;
}
.online-card { background:#e8f5e9; }
.offline-card { background:#fdecea; }

.device-box {
  background:#fff;
  padding:8px;
  border:1px solid #ccc;
  border-radius:6px;
  font-size:12px;
  text-align:center;
  cursor: pointer;
}

/* Main Area */
#main {
  flex:1;
  padding:20px;
  overflow-y:auto;
}

table {
  width:100%;
  border-collapse:collapse;
  margin-top:15px;
  background:white;
}
th, td {
  border:1px solid #ccc;
  padding:6px;
  font-size:13px;
}
th { background:#eee; }
tr:nth-child(even) { background:#fafafa; }

/* Device Panel (right side slide-in) */
#devicePanel {
  position:fixed;
  top:0;
  right:0;
  width:380px;
  height:100vh;
  background:white;
  border-left:1px solid #ccc;
  box-shadow:-2px 0 8px rgba(0,0,0,0.2);
  padding:16px;
  display:none;
  z-index:1000;
  overflow-y:auto;
}

#devicePanelHeader {
  display:flex;
  justify-content:space-between;
  align-items:center;
  margin-bottom:10px;
}

.close-btn {
  cursor:pointer;
  font-weight:bold;
  font-size:18px;
}

/* small table inside panel */
#deviceEventsTable {
  width:100%;
  border-collapse:collapse;
  font-size:12px;
}
#deviceEventsTable th, #deviceEventsTable td {
  border:1px solid #ddd;
  padding:4px;
}
#deviceEventsTable th {
  background:#f0f0f0;
}

</style>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>

<body>

<!-- Sidebar -->
<div id="sidebar">

<h3>Device Status</h3>

<div style="display:flex; gap:10px; margin-bottom:15px;">
  <div class="card online-card" style="flex:1;">
    <b>Online</b><br>
    <span id="onlineCount" style="font-size:22px;"></span>
  </div>

  <div class="card offline-card" style="flex:1;">
    <b>Offline</b><br>
    <span id="offlineCount" style="font-size:22px;"></span>
  </div>
</div>

<h4>Devices</h4>
<div id="healthGrid" style="
  display:grid;
  grid-template-columns:repeat(2,1fr);
  gap:10px;
  margin-bottom:15px;">
</div>

<h4>Severity</h4>
<canvas id="severityChart" width="240" height="180"></canvas>

<h4>Events by Device Type</h4>
<canvas id="deviceTypeChart" width="240" height="180"></canvas>

<h4>Events Timeline</h4>
<canvas id="timelineChart" width="240" height="180"></canvas>

<h4>Device Health Quality</h4>
<canvas id="healthQualityChart" width="240" height="180"></canvas>

</div>

<!-- Main Area -->
<div id="main">

<h2>Mini CDO - Event Dashboard</h2>

<div>
  <button onclick="loadRecent()">Recent</button>

  <input id="devType" placeholder="Device type (FTD/WSA/ESA/ISE)">
  <button onclick="loadByDeviceType()">By Device Type</button>

  <input id="devId" placeholder="Device ID (FTD-1)">
  <button onclick="loadByDeviceID()">By Device ID</button>

  <input id="severity" placeholder="Severity (critical/warning/info)">
  <button onclick="loadBySeverity()">By Severity</button>

  <input id="tenant" placeholder="Tenant">
  <button onclick="loadByTenant()">By Tenant</button>
</div>

<table>
<thead>
<tr>
  <th>Time</th>
  <th>DeviceID</th>
  <th>Device</th>
  <th>Type</th>
  <th>Message</th>
  <th>Details</th>
  <th>Severity</th>
  <th>Tenant</th>
</tr>
</thead>
<tbody id="tbody"></tbody>
</table>

<!-- Device Panel (slide-in) -->
<div id="devicePanel">
  <div id="devicePanelHeader">
    <div>
      <div id="devicePanelTitle" style="font-weight:bold; font-size:16px;"></div>
      <div id="devicePanelStatus" style="font-size:12px; margin-top:2px;"></div>
    </div>
    <div class="close-btn" onclick="closeDevicePanel()">✕</div>
  </div>

  <div id="devicePanelAlerts" style="
       font-size:13px;
       margin-bottom:10px;">
  </div>

  <h4>Recent Events</h4>
  <table id="deviceEventsTable">
    <thead>
      <tr>
        <th>Time</th>
        <th>Type</th>
        <th>Msg</th>
        <th>Severity</th>
      </tr>
    </thead>
    <tbody id="deviceEventsBody"></tbody>
  </table>
</div>

<script>

var severityChartObj = null;
var deviceTypeChartObj = null;
var timelineChartObj = null;
var healthQualityChartObj = null;

// ---------- Health sidebar (online/offline + grid) ----------

async function loadHealth() {
  let r = await fetch("/api/health");
  let d = await r.json();

  let online = 0;
  let offline = 0;
  let html = "";

  for (let dev in d) {
    let st = d[dev];
    if (st === "online") online++; else offline++;

    let color = (st === "online") ? "#28a745" : "#dc3545";

    html += "<div class='device-box' onclick=\\\"openDevicePanel('" + dev + "')\\\">" +
      "<b>"+dev+"</b><br>" +
      "<span style='background:"+color+"; color:white; padding:2px 6px; border-radius:4px;'>"+st+"</span>" +
      "</div>";
  }

  document.getElementById("onlineCount").innerText = online;
  document.getElementById("offlineCount").innerText = offline;
  document.getElementById("healthGrid").innerHTML = html;
}
setInterval(loadHealth, 3000);
loadHealth();

// ---------- Helper to fetch JSON ----------

function fetchJSON(url){
  return fetch(url).then(function(r){ return r.json(); });
}

// ---------- Event table render ----------

function render(events){
  let tb = document.getElementById("tbody");
  tb.innerHTML="";
  events.forEach(function(e){
    tb.innerHTML += "<tr>"+
      "<td>"+(e.time||"")+"</td>"+
      "<td>"+(e.device_id||"")+"</td>"+
      "<td>"+(e.device||"")+"</td>"+
      "<td>"+(e.type||"")+"</td>"+
      "<td>"+(e.msg||"")+"</td>"+
      "<td>"+(e.details||"")+"</td>"+
      "<td>"+(e.severity||"")+"</td>"+
      "<td>"+(e.tenant||"")+"</td>"+
      "</tr>";
  });
}

// ---------- Filters ----------

function loadRecent(){
  fetchJSON("/api/events/recent?limit=200").then(function(ev){
    render(ev);
    loadCharts(ev);
  });
}
function loadByDeviceType(){
  let v=document.getElementById("devType").value;
  fetchJSON("/api/events/by-device-type?device="+v).then(function(ev){
    render(ev);
    loadCharts(ev);
  });
}
function loadByDeviceID(){
  let v=document.getElementById("devId").value;
  fetchJSON("/api/events/by-device-id?device_id="+v).then(function(ev){
    render(ev);
    loadCharts(ev);
  });
}
function loadBySeverity(){
  let v=document.getElementById("severity").value;
  fetchJSON("/api/events/by-severity?severity="+v).then(function(ev){
    render(ev);
    loadCharts(ev);
  });
}
function loadByTenant(){
  let v=document.getElementById("tenant").value;
  fetchJSON("/api/events/by-tenant?tenant="+v).then(function(ev){
    render(ev);
    loadCharts(ev);
  });
}

// ---------- Charts (same as before) ----------

async function loadCharts(events){
  var ev = events;
  if (!ev) {
    ev = await fetchJSON("/api/events/recent?limit=200");
  }

  var info=0, warn=0, crit=0;
  ev.forEach(function(e){
    var s = (e.severity || "").toLowerCase();
    if (s === "info") info++;
    else if (s === "warning") warn++;
    else if (s === "critical") crit++;
  });

  if (severityChartObj) severityChartObj.destroy();
  severityChartObj = new Chart(
    document.getElementById("severityChart").getContext("2d"),
    {
      type: "pie",
      data: {
        labels: ["Info", "Warning", "Critical"],
        datasets: [{
          data: [info, warn, crit],
          backgroundColor: ["#64b5f6", "#ffcc80", "#ef9a9a"]
        }]
      },
      options: {
        plugins: { legend: { display:true, position:"bottom" } }
      }
    }
  );

  var ftd=0, wsa=0, esa=0, ise=0;
  ev.forEach(function(e){
    var dt = (e.device || "").toUpperCase();
    if (!dt && e.device_id) {
      if (e.device_id.indexOf("FTD-") === 0) dt="FTD";
      else if (e.device_id.indexOf("WSA-") === 0) dt="WSA";
      else if (e.device_id.indexOf("ESA-") === 0) dt="ESA";
      else if (e.device_id.indexOf("ISE-") === 0) dt="ISE";
    }
    if (dt === "FTD") ftd++;
    else if (dt === "WSA") wsa++;
    else if (dt === "ESA") esa++;
    else if (dt === "ISE") ise++;
  });

  if (deviceTypeChartObj) deviceTypeChartObj.destroy();
  deviceTypeChartObj = new Chart(
    document.getElementById("deviceTypeChart").getContext("2d"),
    {
      type: "bar",
      data: {
        labels: ["FTD", "WSA", "ESA", "ISE"],
        datasets: [{
          data: [ftd, wsa, esa, ise],
          backgroundColor: "#90caf9"
        }]
      },
      options: {
        plugins: { legend: { display:false } },
        scales: { y: { beginAtZero:true } }
      }
    }
  );

  var timelineMap = {};
  ev.forEach(function(e){
    if (!e.time) return;
    var d = new Date(e.time);
    if (isNaN(d.getTime())) return;
    var hh = (d.getHours()<10 ? "0"+d.getHours() : ""+d.getHours());
    var mm = (d.getMinutes()<10 ? "0"+d.getMinutes() : ""+d.getMinutes());
    var label = hh+":"+mm;
    if (!timelineMap[label]) timelineMap[label]=0;
    timelineMap[label]++;
  });

  var tLabels = Object.keys(timelineMap).sort();
  var tValues = [];
  tLabels.forEach(function(k){ tValues.push(timelineMap[k]); });

  if (timelineChartObj) timelineChartObj.destroy();
  timelineChartObj = new Chart(
    document.getElementById("timelineChart").getContext("2d"),
    {
      type: "line",
      data: {
        labels: tLabels,
        datasets: [{
          data: tValues,
          borderColor: "#42a5f5",
          tension: 0.3
        }]
      },
      options: {
        plugins: { legend: { display:false } },
        scales: { y: { beginAtZero:true } }
      }
    }
  );

  var qualityMap = { good:0, warn:0, bad:0 };
  var h = await fetchJSON("/api/health");
  var devState = {};
  for (var dev in h) {
    if (h[dev] === "online") devState[dev] = "good";
    else devState[dev] = "bad";
  }

  ev.forEach(function(e){
    if (!e.details || !e.device_id) return;
    var cpu = 0, mem = 0;

    var cpuMatch = e.details.match(/cpu=(\d+)/);
    if (cpuMatch && cpuMatch.length > 1) cpu = parseInt(cpuMatch[1]);

    var memMatch = e.details.match(/mem=(\d+)/);
    if (memMatch && memMatch.length > 1) mem = parseInt(memMatch[1]);

    var high = cpu;
    if (mem > high) high = mem;

    if (high >= 90) {
      devState[e.device_id] = "bad";
    } else if (high >= 75) {
      if (devState[e.device_id] !== "bad") {
        devState[e.device_id] = "warn";
      }
    }
  });

  for (var devName in devState) {
    var st = devState[devName];
    if (st === "good") qualityMap.good++;
    else if (st === "warn") qualityMap.warn++;
    else if (st === "bad") qualityMap.bad++;
  }

  if (healthQualityChartObj) healthQualityChartObj.destroy();
  healthQualityChartObj = new Chart(
    document.getElementById("healthQualityChart").getContext("2d"),
    {
      type: "bar",
      data: {
        labels: ["Good", "Warn", "Bad"],
        datasets: [{
          data: [qualityMap.good, qualityMap.warn, qualityMap.bad],
          backgroundColor: ["#66bb6a", "#ffeb3b", "#ef5350"]
        }]
      },
      options: {
        plugins: { legend: { display:false } },
        scales: { y: { beginAtZero:true } }
      }
    }
  );
}

// ---------- Device Panel (right side) ----------

async function openDevicePanel(deviceID) {
  // set title
  document.getElementById("devicePanelTitle").innerText = deviceID;

  // status from /api/health
  let h = await fetchJSON("/api/health");
  let st = h[deviceID] || "unknown";
  let color = (st === "online") ? "green" : (st === "offline" ? "red" : "gray");
  document.getElementById("devicePanelStatus").innerHTML =
    "<span style='color:"+color+"'>Status: "+st+"</span>";

  // alerts from /api/device/alerts
  let alerts = await fetchJSON("/api/device/alerts?device_id="+encodeURIComponent(deviceID));
  let alertsHTML = "";
  if (alerts.length === 0) {
    alertsHTML = "<i>No alerts for this device.</i>";
  } else {
    alertsHTML = "<b>Alerts:</b><ul>";
    alerts.forEach(function(a){
      alertsHTML += "<li>["+a.severity.toUpperCase()+"] "+a.message+" ("+a.time+")</li>";
    });
    alertsHTML += "</ul>";
  }
  document.getElementById("devicePanelAlerts").innerHTML = alertsHTML;

  // events from /api/events/by-device-id
  let ev = await fetchJSON("/api/events/by-device-id?device_id="+encodeURIComponent(deviceID));
  let body = document.getElementById("deviceEventsBody");
  body.innerHTML = "";
  ev.slice(0, 50).forEach(function(e){
    body.innerHTML += "<tr>"+
      "<td>"+(e.time||"")+"</td>"+
      "<td>"+(e.type||"")+"</td>"+
      "<td>"+(e.msg||"")+"</td>"+
      "<td>"+(e.severity||"")+"</td>"+
      "</tr>";
  });

  // show panel
  document.getElementById("devicePanel").style.display = "block";
}

function closeDevicePanel() {
  document.getElementById("devicePanel").style.display = "none";
}

// initial load
loadRecent();
setInterval(function(){ loadCharts(null); }, 15000);

</script>

</div>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
