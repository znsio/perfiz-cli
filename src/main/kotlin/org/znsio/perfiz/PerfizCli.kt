package org.znsio.perfiz

import picocli.CommandLine
import picocli.CommandLine.Command
import java.util.concurrent.Callable
import kotlin.system.exitProcess

@Command(
    name = "perfiz",
    mixinStandardHelpOptions = true,
    version = ["perfiz 0.0.1"],
    description = ["Dockerised Performance Test Setup"]
)
class PerfizCli : Callable<Int> {
    override fun call(): Int {
        println("Starting perfiz...")
        return 0
    }
}

fun main(args: Array<String>): Unit = exitProcess(CommandLine(PerfizCli()).execute(*args))