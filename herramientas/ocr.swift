// OCR con el motor Vision de macOS. Uso: ocr <imagen.png> [...]
// Imprime el texto de cada imagen, en orden, separado por \f (salto de página).
import Foundation
import Vision
import AppKit

let rutas = CommandLine.arguments.dropFirst()
var primero = true
for ruta in rutas {
    if !primero { print("\u{0C}", terminator: "") }
    primero = false
    guard let img = NSImage(contentsOfFile: ruta),
          let cg = img.cgImage(forProposedRect: nil, context: nil, hints: nil) else { continue }
    let req = VNRecognizeTextRequest()
    req.recognitionLevel = .accurate
    req.recognitionLanguages = ["es-ES", "en-US"]
    req.usesLanguageCorrection = true
    try? VNImageRequestHandler(cgImage: cg, options: [:]).perform([req])
    // orden de lectura: de arriba abajo, luego izquierda a derecha
    let obs = (req.results ?? []).sorted {
        let a = $0.boundingBox, b = $1.boundingBox
        if abs(a.midY - b.midY) > 0.012 { return a.midY > b.midY }
        return a.minX < b.minX
    }
    var ultimaY: CGFloat? = nil
    var salida = ""
    for o in obs {
        guard let t = o.topCandidates(1).first?.string else { continue }
        if let y = ultimaY, y - o.boundingBox.midY > 0.035 { salida += "\n" }  // hueco grande = párrafo nuevo
        salida += t + "\n"
        ultimaY = o.boundingBox.midY
    }
    print(salida, terminator: "")
}
