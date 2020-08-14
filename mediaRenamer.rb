#!/usr/bin/ruby -w
# encoding: utf-8
require 'find';
require 'date';
require 'logger';

####################
###Logger =BEGIN=###
####################
timestamp = DateTime.now()
log_file = File.open("MediaRenamer-#{timestamp.strftime("%Y%m%d%H%M%S")}.log", File::WRONLY | File::APPEND | File::CREAT)
logger = Logger.new(log_file)
#logger = Logger.new(STDERR)
#logger = Logger.new(STDOUT)
$stderr = log_file
#$stdout = log_file
logger.level = Logger::DEBUG
logger.datetime_format = '%Y-%m-%d %H:%M:%S '
logger.formatter = proc do |severity, datetime, progname, msg|
  "#{severity}: #{msg}\n"
end
##################
###Logger =END=###
##################

logger.info("Старт програми: #{timestamp.strftime("%Y-%m-%d %H:%M:%S")}")
files_all = 0 #для підрахунку всіх файлів
files_renamed = 0
files_sciped = 0
files_error = Array.new()
no_exif_data = Array.new()
puts ""
puts "-----------------------MediaRenamer v. Alpha-------------------------------"
puts "Програма працює з файлами типу: JPG, ARW, NEF, AVI, 3GP, MP4, M4V, MOV та MTS."
if File.exist?("C:/bin/exiftool.exe") or File.exist?('/usr/bin/exiftool')
    puts "Exiftool знайдено."
    if File.exist?("C:/bin/exiftool.exe")
        exiftool = "C:/bin/exiftool.exe"
    else
        exiftool = "/usr/bin/exiftool"
    end
    logger.info {"Exiftool знайдено: #{exiftool}"}
else
    print "Exiftool не знайдено, введіть yes якщо бажаєте опрацювати тільки файли що містять дату в імені: "
    answer = gets.downcase.chomp
    logger.warn("Exiftool не знайдено. Відповідь користувача на запит продовження: #{answer}")
    if (answer =~ /^yes$/i)
        puts "Будуть оброблені тільки JPG файли що містять дату в імені."
    else
        logger.close
        exit
    end
end
if ARGV[0].nil?
	print "Пропишіть теку, яку бажаєте обробити, або ENTER щоб обробити поточну теку:  "
	where = gets.chomp
	if where.length == 0
		where = Dir.pwd.encode("UTF-8")
	end
else
	where = ARGV[0].encode("UTF-8")
end
puts "----------------------------------------------------------------------"
logger.info("Пошук починаючи з папки " + where + "\n==============================") and puts "Пошук починаючи з папки " + where
puts "----------------------------------------------------------------------"
puts ""

#####################
###RENAMER =BEGIN=###
#####################
renamer = lambda {|newname, path, mark|
    counter = 1
    full_path = "#{File.absolute_path(File.dirname(path))}" + "/"
    extension = File.extname(path).downcase
    if File.exist?("#{full_path}#{newname}#{extension}")
        logger.info("Файл із іменем #{newname}#{extension} існує, додаю лічильник:")
        counter += 1
        logger.info(counter)
        while File.exist?("#{full_path}#{newname}_(#{counter.to_s})#{extension}") do
             counter += 1
             logger.info(counter)
        end
    end
    if counter > 1
             newname = newname.to_s + "_(" + counter.to_s + ")"
             logger.info("Нове ім’я: " + newname)
    end
    if File.rename(path, "#{full_path}#{newname}#{extension}")
        logger.info("Перейменовую у " + "#{newname}#{extension}. #{mark}\n=====")
    else
        logger.error("Перейменування не вдалось!\n-----------!!!ERROR!!!-----------\n--->>>#{path}\n=====")
        path
    end
    }
###################
###RENAMER =END=###
###################

##################
###FIND =BEGIN=###
##################
Find.find(where) do |path|
  path = path.encode("UTF-8")
  if FileTest.directory?(path)
    if File.basename(path)[0] == ?. # Якщо треба виключити теку з обробки, перейменуйте назву теки щоб спочатку була крапка у назві. Наприклад => .ФотоТЕКА
		logger.info(path + " Тека має крапку першим знаком у назві - оминається!\n=====")
      Find.prune       # Don't look any further into this directory.
      logger.debug(path + "оминається.")
    else
      next
    end
  else
    ext = File.extname(path).downcase
	if ext =~ /^.*\.jpg$|^.*\.mts$|^.*\.mp4$|^.*\.3gp$|^.*\.arw|^.*\.m4v$|^.*\.mov$|^.*\.avi$|^.*\.png|^.*\.nef$|^.*\.cr2$|^.*\.jpeg$/i
      puts path
        case
        when File.basename(path)[0] == ?. # Файл який має першу крапку у назві - оминається
            files_all += 1
            files_sciped += 1
            logger.info("Файл: " +  path + ". Має першу крапку у назві - оминається!\n=====")
            next
        when (File.basename(path, ".*") =~ /^\d{8}_\d{6}$/i) || (File.basename(path, ".*") =~ /^\d{8}_\d{6}\(\d+\)$/i) || (File.basename(path, ".*") =~ /^\d{8}_\d{6}_\(\d+\)$/i) # якщо файл має ім’я 20141012_162920 або 20141012_162920(2)
            files_all += 1
            files_sciped += 1
            logger.info("Файл: " +  path + ". Не потребує перейменування!\n=====")
        when File.basename(path, ".*") =~ /^[A-Z]{3}_\d{8}_\d{6}$/i
            files_all += 1
            files_renamed += 1
            logger.info("Файл: " +  path)
            logger.debug('===> =~ /^[A-Z_]{3}_\d{8}_\d{6}$/i')
            newname = /\d{8}_\d{6}$/.match(File.basename(path, ".*"))
            result = renamer.call(newname, path, 'По імені.')
            unless result.nil? or result
                files_error.push(result)
            end
        when (File.basename(path, ".*") =~ /^.*\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}.*$/i)
            files_all += 1
            files_renamed += 1
            logger.info("Файл: " +  path + ' ===>/^.*\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}.*$/i')
            logger.debug('===> =~ /^.*\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}.*$/i')
            data = /^.*(\d{4})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2}).*$/i.match(path)
            newname = "#{Regexp.last_match[1].to_s}#{Regexp.last_match[2].to_s}#{Regexp.last_match[3].to_s}_#{Regexp.last_match[4].to_s}#{Regexp.last_match[5].to_s}#{Regexp.last_match[6].to_s}"
            result = renamer.call(newname, path, 'По імені.')
            unless result.nil? or result
                files_error.push(result)
            end
        when !exiftool.nil?
            files_all += 1
            files_renamed += 1
            logger.info("Файл: " +  path )
            logger.debug(' ==>!exiftool.nil?')
            info = `#{exiftool} -b -d \"%Y%m%d_%H%M%S\" -DateTimeOriginal \"#{path}\"`.downcase.chomp
            # info = `#{exiftool} -b -d \"%Y%m%d_%H%M%S\" -ModifyDate \"#{path}\"`.downcase.chomp
            case
            when info.nil?, info.eql?('0000:00:00 00:00:00'), info.empty?, info.length < 15
            	logger.debug("Exif недійсний ===>" + info.to_s + "<=== Передається для обробки по даті ФС.\n=====")
            	no_exif_data.push(path.chomp)
                files_all -= 1
                files_renamed -= 1
                next
			when info[0..3].to_i < 2005 || info[0..3].to_i == 0
				logger.debug("Exif недійсний ===>" + info.to_s + "<=== Передається для обробки по даті ФС.\n=====")
				no_exif_data.push(path.chomp)
				files_all -= 1
				files_renamed -= 1
				next
            when info =~ /\d{8}_\d{6}/i
            	newname = info
            when info =~ /(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})/i
            	data = /(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2}):(\d{2})/i.match(info)
            	newname = "#{Regexp.last_match[1].to_s}#{Regexp.last_match[2].to_s}#{Regexp.last_match[3].to_s}_#{Regexp.last_match[4].to_s}#{Regexp.last_match[5].to_s}#{Regexp.last_match[6].to_s}"
            end
            result = renamer.call(newname, path, 'По висновку exiftool.')
            unless result.nil? or result
                files_error.push(result)
            end
        end
    else
        next
	end
  end
end

################
###FIND =END=###
################

unless no_exif_data.empty?
    no_exif_data.each {|path|
        files_all += 1
        files_renamed += 1
        logger.info("Filename: " +  path)
        data = File.mtime(path)
        newname = data.strftime("%Y%m%d_%H%M%S")
        result = renamer.call(newname, path, 'По даті ФС.')
        unless result.nil? or result
            files_error.push(result)
        end
    }
end
if files_all >= 0
    logger.info("Всього файлів знайдено: #{files_all}.")
    logger.info("В тому числі:")
    logger.info("\t перейменовано: #{files_renamed}.")
    logger.info("\t ім’я не змінено: #{files_sciped}.")
    unless files_error.empty?
        logger.info("#{files_error.length} файлів не вдалось обробити:")
        logger.info("#{puts files_error}")
    end
else
    logger.info("Не знайдено жодного файла.")
end
logger.info("Завершення програми: #{DateTime.now().strftime("%Y-%m-%d %H:%M:%S")}")
logger.close
