/* globals HTMLElement, customElements */
import L from 'leaflet'

class App extends HTMLElement {
  #markers = []
  constructor () {
    super()
    this.attachShadow({ mode: 'open' })
    this.shadowRoot.innerHTML = `
        <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" integrity="sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=" crossorigin=""/>
        <style>
            :host {
                position: absolute;
                display: block;
                inset: 0;
                width: 100dvw;
                height: 100dvh;

                #map {
                    position: absolute;
                    inset: 0;
                }
            }
        </style>
        <div id="map"></div>
    `
  }

  async connectedCallback () {
    // Setup Vancouver island map
    this.map = new L.Map(this.shadowRoot.querySelector('#map')).setView([49.499998, -125.499998], 8)
    const bounds = L.latLngBounds(
      L.latLng(48.004625, -129.264247),
      L.latLng(51.371780, -122.482797)
    )
    this.map.fitBounds(bounds)

    // Setup Canada Map
    // const map = new L.Map('map').setView([63.770344, -96.81157], 4)
    // const bounds = L.latLngBounds(
    //   latLng(76.67978490310692, -13.798828125000002),
    //   latLng(40.111688665595956, -179.912109375)
    // )

    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
    }).addTo(this.map)

    this.map.on('moveend', this.handleEvent.bind(this))

    const points = await this.fetchPoints(bounds)
    this.markPoints(points)
  }

  async handleEvent (e) {
    const bounds = this.map.getBounds()
    const points = await this.fetchPoints(bounds)
    this.markPoints(points)
  }

  async fetchPoints (bounds) {
    const res = await fetch(`http://localhost:4000/api/v1/?coords=${bounds.toBBoxString()}`)
    const body = await res.json()
    const data = body.data

    return data
  }

  markPoints (points) {
    for (const oldMarker of this.#markers) {
      oldMarker.remove()
    }

    if (!points?.length) return
    for (const entry of points) {
      const marker = this.createMarker(entry)
      const toolTip = this.createToolTip(entry)

      this.placeMarker(marker, toolTip)
      this.#markers.push(marker)
    }
  }

  createMarker (point) {
    const marker = L.marker([point.latitude, point.longitude], {
      title: 'test',
      opacity: point.magnitude / 10,
    })

    return marker
  }

  createToolTip ({ title, elevation, latitude, longitude, magnitude }) {
    const details = {
      elevation,
      latitude,
      longitude,
      magnitude,
    }

    return `<strong>${title}</strong><hr><ul>${Object.entries(details).map(([key, value]) => `<li>${key}: ${value}</li>`).join('\n')}</ul>`
  }

  placeMarker (marker, toolTip) {
    marker.addTo(this.map).bindTooltip(toolTip)
  }

  /// Returns a scaled copy of the marker's icon.
  scaleIconForMarker (marker, enlargeFactor) {
    const iconOptions = marker.options.icon.options

    const newIcon = L.icon({
      iconUrl: iconOptions.iconUrl,
      iconSize: this.multiplyPointBy(enlargeFactor, iconOptions.iconSize),
      iconAnchor: this.multiplyPointBy(enlargeFactor, iconOptions.iconAnchor),
    })

    return newIcon
  }

  /// Helper function, for some reason,
  /// Leaflet's Point.multiplyBy(<Number> num) function is not working for me,
  /// so I had to create my own version, lol
  /// refer to https://leafletjs.com/reference.html#point-multiplyby
  multiplyPointBy (factor, originalPoint) {
    const newPoint = L.point(
      originalPoint[0] * factor,
      originalPoint[1] * factor
    )

    return newPoint
  }
}

export const register = () => customElements.define('quake-map', App)
