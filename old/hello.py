import subprocess
from flask import Flask, request, jsonify, flash, redirect, url_for, send_file
from flask_cors import CORS, cross_origin
import pyodbc
from werkzeug.utils import secure_filename
import os
import base64
# import zipfile38 as zipfile


UPLOAD_FOLDER = './uploads'
ALLOWED_EXTENSIONS = set(['tif', 'png', 'jpg', 'jpeg'])

app = Flask(__name__)
#CORS(app)
CORS(app, resources={r"/*": {"origins": "*"}})

app.secret_key = "super secret key"
app.config.from_pyfile('config.py')
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
# app.config['CORS_HEADERS'] = 'Content-Type'

# this check makes sure there is a uploads folder in place.
CHECK_UPLOAD_FOLDER = os.path.isdir(app.config['UPLOAD_FOLDER'])
# print("folder exists: " + str(CHECK_UPLOAD_FOLDER))
if not CHECK_UPLOAD_FOLDER:
    os.makedirs(app.config['UPLOAD_FOLDER'])
    # print("created folder : ", app.config['UPLOAD_FOLDER'])



@app.after_request
def after_request(response):
    # response.headers.add('Access-Control-Allow-Origin', 'https://hdr-generator.azurewebsites.net')
    # response.headers.add('Access-Control-Allow-Origin', '*')
    response.headers.add('Access-Control-Allow-Headers', 'Content-Type,Authorization')
    response.headers.add('Access-Control-Allow-Methods', 'GET,PUT,POST,DELETE,OPTIONS')
    return response

server = app.config['SQL_SERVER']
database = app.config['SQL_DATABASE']
username = app.config['SQL_USERNAME']
password = app.config['SQL_PASSWORD']
driver = app.config['SQL_DRIVER']

# set script names
CAMGEN = "0_camgen.csh"
RUNHDR = "1_runhdr"
UPEXPOSE = "11_upexpose"
DOWNEXPOSE = "12_downexpose"
MATRIX = "2_matrix"
SCALING = "21_scaling"
IMAGES = "3_images"




def allowed_file(filename):
    return '.' in filename and filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS


# @app.route('/api/downloadBulk')
# def downloadBulk():
#     pn = request.args.get('pn')
#     if not pn:
#         return jsonify({ 'response': "Error, pn required" }), 400
    
#     zipf = zipfile.ZipFile(f'./uploads/{str(pn)}/tif/{pn}_HDR_downloads.zip','w', zipfile.ZIP_DEFLATED)
#     files = []

#     for root, dirs, files in os.walk(f'./uploads/{str(pn)}/tif/'):
#         files = [ fi for fi in files if fi.endswith(".jpg") ]
    
#     for file in files:
#         zipf.write(f'./uploads/{str(pn)}/tif/' + file)
#     zipf.close()

#     return send_file(f'./uploads/{str(pn)}/tif/{pn}_HDR_downloads.zip',
#             mimetype = 'zip',
#             attachment_filename= f'./uploads/{str(pn)}/tif/{pn}_HDR_downloads.zip',
#             as_attachment = True)
    

@app.route('/api/uploadImage', methods=['POST', 'OPTIONS'])
@cross_origin()
def upload_image():
    pn = request.form['pn'] 
    img_type = 'originals'
    img_name = request.form['img_name']
    test = request.form['test']

    if not img_name or not img_type or not pn:
        return "Error: pn, img_type and img_name required"

    if 'file' not in request.files:
        return jsonify({ 'response': 'No file part' }), 400
    
    files = request.files.getlist("file")

    if test == 'false':
        for file in files:
            print(file.filename)
            print("___________")

            if file.filename == '':
                return jsonify({ 'response': 'No selected file' }), 400
            
            CHECK_FOLDER = os.path.isdir(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals' + '/VP_' + str(img_name))
            if not CHECK_FOLDER:
                os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals' + '/VP_' + str(img_name))
                print("created folder : ", app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals')

            if file and allowed_file(file.filename):
                filename = secure_filename(file.filename)
                file.save(os.path.join(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/' + str(img_type) + '/VP_' + str(img_name), filename))
    else:
        for file in files:
            print(file.filename)
            print("___________")

            if file.filename == '':
                return jsonify({ 'response': 'No selected file' }), 400
            
            CHECK_FOLDER = os.path.isdir(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals')
            if not CHECK_FOLDER:
                os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals')
                print("created folder : ", app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals')

            if file and allowed_file(file.filename):
                filename = secure_filename(file.filename)
                file.save(os.path.join(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/' + str(img_type), filename))

    try:
        with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
            with conn.cursor() as cursor:
                cursor.execute("""INSERT INTO dbo.HDR_Images 
                        (Project_Number,Image_Type,Image_Url,Image_Name ) 
                        VALUES (""" 
                        + request.form['pn'] 
                        + ", '" + str(img_type)
                        + "', '" + "not active" 
                        + "', '" + str(img_name) 
                        + "')")

        return jsonify({ 'response': 'Image set added to DB' }), 201
    except:
        return jsonify({ 'response': 'Image set with this name already exists' }), 200



@app.route('/api/checkCamgen')
def checkCamgen():
    pn = request.args.get('pn')
    if not pn:
        return jsonify({ 'response': "Error, pn required" }), 400
    
    checkFile = os.path.isfile(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/yourcamera.cam')
    if checkFile:
        return jsonify({ 'response': '.cam file exists' }), 200
    else:
        return jsonify({ 'response': '.cam file does not exist' }), 200
    


@app.route('/api/deleteCamgen')
def deleteCamgen():
    pn = request.args.get('pn')
    if not pn:
        return jsonify({ 'response': "Error, pn required" }), 400
    
    subprocess.call("rm ./yourcamera.cam", cwd=f"./uploads/{str(pn)}/", shell=True)
    return jsonify({ 'response': 'camgen file deleted' }), 200


@app.route('/api/getImageByName')
def getImageByName():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    if not pn or not img_name:
        return jsonify({ 'response': "Error, pn and img required" }), 400
    
    checkFile = os.path.isfile(f"./uploads/{str(pn)}/tif/{str(img_name)}.tif")
    if not checkFile:
        return jsonify({ 'response': 'file does not exist' }), 200
    else:
        subprocess.call(f"convert ./uploads/{str(pn)}/tif/{str(img_name)}.tif ./uploads/{str(pn)}/tif/{str(img_name)}.jpg", shell=True)
        data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.jpg', 'rb').read()
        bytes_base64 = base64.b64encode(data)
        text_base64 = bytes_base64.decode()
        return jsonify({ 'image': text_base64 }), 200


@app.route('/api/upexposeImage')
def upexposeImage():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    factor = str(request.args.get('factor'))

    if not pn or not img_name:
        return jsonify({ 'response': "Error, pn and img required" }), 400
    # call script
    subprocess.call(f"./{UPEXPOSE} {str(img_name)}.comb {factor}", cwd=f"./uploads/{str(pn)}/", shell=True)
    subprocess.call(f"convert ./tif/{str(img_name)}.comb.tif ./tif/{str(img_name)}.comb.jpg", cwd=f"./uploads/{str(pn)}/", shell=True)
    data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.comb.jpg', 'rb').read()
    
    bytes_base64 = base64.b64encode(data)
    text_base64 = bytes_base64.decode()
    return jsonify({ 'image': text_base64 }), 200


@app.route('/api/downexposeImage')
def downexposeImage():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    factor = str(request.args.get('factor'))

    if not pn or not img_name:
        return jsonify({ 'response': "Error, pn and img required" }), 400
    # call script
    subprocess.call(f"./{DOWNEXPOSE} {str(img_name)}.comb {factor}", cwd=f"./uploads/{str(pn)}/", shell=True)
    subprocess.call(f"convert ./tif/{str(img_name)}.comb.tif ./tif/{str(img_name)}.comb.jpg", cwd=f"./uploads/{str(pn)}/", shell=True)
    data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.comb.jpg', 'rb').read()
    
    bytes_base64 = base64.b64encode(data)
    text_base64 = bytes_base64.decode()
    return jsonify({ 'image': text_base64 }), 200
    
@app.route('/api/processCamgen')
def processCamgen():
    pn = request.args.get('pn')
    if not pn:
        return jsonify({ 'response': "Error, pn required" }), 400

    subprocess.call(f"./{CAMGEN}", cwd=f"/home/airflow-machine-admin/backendHDR/uploads/{str(pn)}/", shell=True)
    return jsonify({ 'response': 'camgen processing success' }), 200

@app.route('/api/deleteProject')
def deleteProject():
    pn = request.args.get('pn')

    if not pn:
        return jsonify({ 'response': "Error, pn required" }), 400
    else:
        # delete from VM
        subprocess.call(f"rm -rf ./{str(pn)}", cwd="./uploads/", shell=True)
        # delete from Image DB
        try:
            with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
            	with conn.cursor() as cursor:
                    cursor.execute(f"DELETE FROM dbo.HDR_Images WHERE Project_Number = {pn}")
        
        except Exception as e:
            print(e)
            return jsonify({ 'response': 'Error deleting project images.' }), 200

        # delete from Project DB
        try:
            with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
            	with conn.cursor() as cursor:
                    cursor.execute(f"DELETE FROM dbo.HDR_Projects WHERE Project_Number = {pn}")
            return jsonify({ 'response': f'Project {pn} Deleted' }), 200
        
        except Exception as e:
            print(e)
            return jsonify({ 'response': 'Error deleting project or it may not exist.' }), 200



@app.route('/api/processHDR')
@cross_origin()
def processHDR():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    if not pn or not img_name:
        # return "Error, pn and img required"
        return jsonify({ 'response': "Error, pn and img required" }), 400

    subprocess.call(f"./{RUNHDR} {img_name}" , cwd=f"/home/airflow-machine-admin/backendHDR/uploads/{str(pn)}/", shell=True)
    subprocess.call(f"convert ./tif/{str(img_name)}.comb.tif ./tif/{str(img_name)}.comb.jpg", cwd=f"/home/airflow-machine-admin/backendHDR/uploads/{str(pn)}/", shell=True)
    data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.comb.jpg', 'rb').read()
    
    bytes_base64 = base64.b64encode(data)
    text_base64 = bytes_base64.decode()
    return jsonify({ 'image': text_base64 }), 200


@app.route('/api/processMatrix')
def processMatrix():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    if not pn or not img_name:
        return jsonify({ 'response': "Error, pn and img required" }), 400

    subprocess.call(f"./{MATRIX} {img_name}", cwd=f"./uploads/{str(pn)}/", shell=True)
    subprocess.call(f"convert ./tif/{str(img_name)}.vischeck.tif ./tif/{str(img_name)}.vischeck.jpg", cwd=f"./uploads/{str(pn)}/", shell=True)
    data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.vischeck.jpg', 'rb').read()
    
    bytes_base64 = base64.b64encode(data)
    text_base64 = bytes_base64.decode()
    return jsonify({ 'image': text_base64 }), 200



@app.route('/api/processScaling')
def processScaling():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    previous = str(request.args.get('previous'))
    target = str(request.args.get('target'))

    if not pn or not img_name or not previous or not target:
        return jsonify({ 'response': "Error, pn, img_name, previous and target required" }), 400

    # requires read state of "1" factor
    # result = str((float(previous) / float(target)) * 1)
    result = str((float(target) / float(previous) ) * 1)

    subprocess.call(f"./{SCALING} {img_name} {result}", cwd=f"./uploads/{str(pn)}/", shell=True)
    subprocess.call(f"convert ./tif/{str(img_name)}.vischeck.tif ./tif/{str(img_name)}.vischeck.jpg", cwd=f"./uploads/{str(pn)}/", shell=True)
    data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.vischeck.jpg', 'rb').read()
    
    bytes_base64 = base64.b64encode(data)
    text_base64 = bytes_base64.decode()
    return jsonify({ 'image': text_base64 }), 200


@app.route('/api/processFalseColor')
def processFalseColor():
    pn = request.args.get('pn')
    img_name = str(request.args.get('img_name'))
    if not pn or not img_name:
        return jsonify({ 'response': "Error, pn and img_name required" }), 400

    subprocess.call(f"./{IMAGES} {img_name}", cwd=f"./uploads/{str(pn)}/", shell=True)

    subprocess.call(f"convert ./tif/{str(img_name)}.fc.100.tif ./tif/{str(img_name)}.fc.100.jpg", cwd=f"./uploads/{str(pn)}/", shell=True)
    data = open(f'./uploads/{str(pn)}/tif/{str(img_name)}.fc.100.jpg', 'rb').read()
    
    bytes_base64 = base64.b64encode(data)
    text_base64 = bytes_base64.decode()
    return jsonify({ 'image': text_base64 }), 200

@app.route('/api/')
def index():
    return jsonify({ '''Welcome to the LVA Radiance instance <br><br>
            go to /help for more info''' }), 200



@app.route('/api/help')
def help():
    return '''intentionally blank''', 200



@app.route('/api/getProjects')
def getProjects():
    with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
        with conn.cursor() as cursor:
            cursor.execute("SELECT * FROM HDR_projects")
            results = cursor.fetchall()
            fields_list = cursor.description
            print(fields_list)
            response = []
            
            for i in results:
                data = {
                    fields_list[0][0]:i[0],
                    fields_list[1][0]:i[1],
                    fields_list[2][0]:i[2]
                }
                response.append(data)

    return jsonify(response), 200




@app.route('/api/createProject', methods=['POST', 'OPTIONS'])
@cross_origin()
def createProject():

    pn = request.form['pn']
    CHECK_FOLDER = os.path.isdir(app.config['UPLOAD_FOLDER'] + '/' + str(pn))
    if not CHECK_FOLDER:
        # create all the required folders
        os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/originals')
        os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/exif')
        os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/pic')
        os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/tif')
        os.makedirs(app.config['UPLOAD_FOLDER'] + '/' + str(pn) + '/tmp')
        print("created folder : ", app.config['UPLOAD_FOLDER'] + '/' + str(pn))

        # copy all the radiance scripts into the project folder and make each executable
        subprocess.call(f"cp ./scripts/{CAMGEN} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{CAMGEN}", shell=True)
        subprocess.call(f"cp ./scripts/{RUNHDR} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{RUNHDR}", shell=True)
        subprocess.call(f"cp ./scripts/{UPEXPOSE} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{UPEXPOSE}", shell=True)
        subprocess.call(f"cp ./scripts/{DOWNEXPOSE} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{DOWNEXPOSE}", shell=True)
        subprocess.call(f"cp ./scripts/{MATRIX} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{MATRIX}", shell=True)
        subprocess.call(f"cp ./scripts/{SCALING} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{SCALING}", shell=True)
        subprocess.call(f"cp ./scripts/{IMAGES} ./uploads/{str(pn)}/", shell=True)
        subprocess.call(f"chmod +x ./uploads/{str(pn)}/{IMAGES}", shell=True)

    try:
        with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
            with conn.cursor() as cursor:
                cursor.execute("""INSERT INTO dbo.HDR_Projects 
                        (Project_Number,Project_Name,Project_Location ) 
                        VALUES (""" 
                        + request.form['pn'] 
                        + ", '" + request.form['Project_Name'] 
                        + "', '" + request.form['Project_Location'] 
                        + "')")

        return jsonify({ 'response': 'Project created' }), 201
    except:
        return jsonify({ 'response': 'Project already exists' }), 200



@app.route('/api/getImages')
def getImages():
    with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
        with conn.cursor() as cursor:
            cursor.execute("SELECT * FROM HDR_Images")
            results = cursor.fetchall()
            fields_list = cursor.description
            print(fields_list)
            response = []
            
            for i in results:
                data = {
                    fields_list[0][0]:i[0],
                    fields_list[1][0]:i[1],
                    fields_list[2][0]:i[2],
                    fields_list[3][0]:i[3],
                    fields_list[4][0]:i[4]
                }
                response.append(data)
    return jsonify(response), 200



@app.route('/api/getProjectImages')
def getProjectImages():
    pn = request.args.get('pn')

    if not pn:
        return jsonify({ 'response': "Error, pn required" }), 400

    with pyodbc.connect('DRIVER='+driver+';SERVER='+server+';PORT=1433;DATABASE='+database+';UID='+username+';PWD='+ password) as conn:
        with conn.cursor() as cursor:
            cursor.execute("SELECT * FROM HDR_Images WHERE Project_Number = " + str(pn))
            results = cursor.fetchall()
            fields_list = cursor.description
            response = []
            
            for i in results:
                data = {
                    fields_list[0][0]:i[0],
                    fields_list[1][0]:i[1],
                    fields_list[2][0]:i[2],
                    fields_list[3][0]:i[3],
                    fields_list[4][0]:i[4]
                }
                response.append(data)

    return jsonify(response), 200

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()
    
@app.route('/shutdown', methods=['GET'])
def shutdown():
    shutdown_server()
    return 'Server shutting down...'


if __name__ == '__main__':
    #app.run(debug=True, host="0.0.0.0", port=5000, ssl_context=('cert.pem', 'key.pem'))
    #app.run(debug=True, host="0.0.0.0", port=5000)
    app.run(host="0.0.0.0", port=5000)
